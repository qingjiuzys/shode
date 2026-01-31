// Package debug 提供性能分析功能。
package debug

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

// Profiler 性能分析器
type Profiler struct {
	cpuProfile    *os.File
	memProfile    *os.File
	blockProfile  *os.File
	mutexProfile  *os.File
	running       bool
	mu            sync.Mutex
	samplingRate  int
}

// ProfileConfig 分析配置
type ProfileConfig struct {
	CPUProfile    string
	MemProfile    string
	BlockProfile  string
	MutexProfile  string
	SamplingRate  int // CPU 采样率
}

// NewProfiler 创建性能分析器
func NewProfiler() *Profiler {
	return &Profiler{
		running:      false,
		samplingRate: 100, // 默认 100Hz
	}
}

// StartCPUProfile 开始 CPU 性能分析
func (p *Profiler) StartCPUProfile(filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cpuProfile != nil {
		return fmt.Errorf("CPU profiling already started")
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create CPU profile: %w", err)
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		return fmt.Errorf("failed to start CPU profile: %w", err)
	}

	p.cpuProfile = f
	fmt.Printf("CPU profiling started, writing to %s\n", filename)
	return nil
}

// StopCPUProfile 停止 CPU 性能分析
func (p *Profiler) StopCPUProfile() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cpuProfile == nil {
		return fmt.Errorf("CPU profiling not started")
	}

	pprof.StopCPUProfile()
	if err := p.cpuProfile.Close(); err != nil {
		return fmt.Errorf("failed to close CPU profile: %w", err)
	}

	fmt.Println("CPU profiling stopped")
	p.cpuProfile = nil
	return nil
}

// WriteMemProfile 写入内存分析
func (p *Profiler) WriteMemProfile(filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create memory profile: %w", err)
	}

	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		return fmt.Errorf("failed to write heap profile: %w", err)
	}

	fmt.Printf("Memory profile written to %s\n", filename)
	return nil
}

// StartBlockProfile 开始阻塞分析
func (p *Profiler) StartBlockProfile(rate int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	runtime.SetBlockProfileRate(rate)
	fmt.Printf("Block profiling started with rate %d\n", rate)
	return nil
}

// StopBlockProfile 停止阻塞分析
func (p *Profiler) StopBlockProfile() {
	p.mu.Lock()
	defer p.mu.Unlock()

	runtime.SetBlockProfileRate(0)
	fmt.Println("Block profiling stopped")
}

// WriteBlockProfile 写入阻塞分析
func (p *Profiler) WriteBlockProfile(filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.blockProfile != nil {
		p.blockProfile.Close()
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create block profile: %w", err)
	}

	defer f.Close()

	if err := pprof.Lookup("block").WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write block profile: %w", err)
	}

	p.blockProfile = f
	fmt.Printf("Block profile written to %s\n", filename)
	return nil
}

// StartMutexProfile 开始互斥锁分析
func (p *Profiler) StartMutexProfile(rate int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	runtime.SetMutexProfileFraction(rate)
	fmt.Printf("Mutex profiling started with fraction %d\n", rate)
	return nil
}

// StopMutexProfile 停止互斥锁分析
func (p *Profiler) StopMutexProfile() {
	p.mu.Lock()
	defer p.mu.Unlock()

	runtime.SetMutexProfileFraction(0)
	fmt.Println("Mutex profiling stopped")
}

// WriteMutexProfile 写入互斥锁分析
func (p *Profiler) WriteMutexProfile(filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.mutexProfile != nil {
		p.mutexProfile.Close()
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create mutex profile: %w", err)
	}

	defer f.Close()

	if err := pprof.Lookup("mutex").WriteTo(f, 0); err != nil {
		return fmt.Errorf("failed to write mutex profile: %w", err)
	}

	p.mutexProfile = f
	fmt.Printf("Mutex profile written to %s\n", filename)
	return nil
}

// MemoryStats 内存统计
type MemoryStats struct {
	Alloc         uint64
	TotalAlloc    uint64
	Sys           uint64
	HeapAlloc     uint64
	HeapSys       uint64
	HeapIdle      uint64
	HeapInuse     uint64
	HeapReleased  uint64
	HeapObjects   uint64
	StackInuse    uint64
	StackSys      uint64
	MSpanInuse    uint64
	MSpanSys      uint64
	MCacheInuse   uint64
	MCacheSys     uint64
	BuckHashSys   uint64
	GCSys         uint64
	OtherSys      uint64
	NextGC        uint64
	LastGC        uint64
	PauseTotalNs  uint64
	NumGC         uint32
	NumForcedGC   uint32
	GCCPUFraction float64
}

// GetMemoryStats 获取内存统计
func (p *Profiler) GetMemoryStats() MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemoryStats{
		Alloc:         m.Alloc,
		TotalAlloc:    m.TotalAlloc,
		Sys:           m.Sys,
		HeapAlloc:     m.HeapAlloc,
		HeapSys:       m.HeapSys,
		HeapIdle:      m.HeapIdle,
		HeapInuse:     m.HeapInuse,
		HeapReleased:  m.HeapReleased,
		HeapObjects:   m.HeapObjects,
		StackInuse:    m.StackInuse,
		StackSys:      m.StackSys,
		MSpanInuse:    m.MSpanInuse,
		MSpanSys:      m.MSpanSys,
		MCacheInuse:   m.MCacheInuse,
		MCacheSys:     m.MCacheSys,
		BuckHashSys:   m.BuckHashSys,
		GCSys:         m.GCSys,
		OtherSys:      m.OtherSys,
		NextGC:        m.NextGC,
		LastGC:        m.LastGC,
		PauseTotalNs:  m.PauseTotalNs,
		NumGC:         m.NumGC,
		NumForcedGC:   m.NumForcedGC,
		GCCPUFraction: m.GCCPUFraction,
	}
}

// PrintMemoryStats 打印内存统计
func (p *Profiler) PrintMemoryStats() {
	stats := p.GetMemoryStats()

	fmt.Println("Memory Statistics:")
	fmt.Printf("  Alloc:        %v MB\n", stats.Alloc/1024/1024)
	fmt.Printf("  TotalAlloc:   %v MB\n", stats.TotalAlloc/1024/1024)
	fmt.Printf("  Sys:          %v MB\n", stats.Sys/1024/1024)
	fmt.Printf("  HeapAlloc:    %v MB\n", stats.HeapAlloc/1024/1024)
	fmt.Printf("  HeapSys:      %v MB\n", stats.HeapSys/1024/1024)
	fmt.Printf("  HeapObjects:  %d\n", stats.HeapObjects)
	fmt.Printf("  StackInuse:   %v MB\n", stats.StackInuse/1024/1024)
	fmt.Printf("  StackSys:     %v MB\n", stats.StackSys/1024/1024)
	fmt.Printf("  NumGC:        %d\n", stats.NumGC)
	fmt.Printf("  PauseTotalNs: %v ms\n", stats.PauseTotalNs/1000000)
}

// GCStats GC 统计
type GCStats struct {
	LastGC         time.Time
	NumGC          uint32
	PauseTotal     time.Duration
	Pause          []time.Duration
	PauseEnd       []time.Time
	NextGC         uint64
	NumForcedGC    uint32
	GCCPUFraction  float64
	EnableGC       bool
	DebugGC        bool
}

// GetGCStats 获取 GC 统计
func (p *Profiler) GetGCStats() GCStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	pauses := make([]time.Duration, len(m.Pause))
	pauseEnds := make([]time.Time, len(m.PauseEnd))

	for i, pauseNs := range m.Pause {
		pauses[i] = time.Duration(pauseNs) * time.Nanosecond
		pauseEnds[i] = time.Unix(0, int64(m.PauseEnd[i]))
	}

	return GCStats{
		LastGC:         time.Unix(0, int64(m.LastGC)),
		NumGC:          m.NumGC,
		PauseTotal:     time.Duration(m.PauseTotalNs) * time.Nanosecond,
		Pause:          pauses,
		PauseEnd:       pauseEnds,
		NextGC:         m.NextGC,
		NumForcedGC:    m.NumForcedGC,
		GCCPUFraction:  m.GCCPUFraction,
		EnableGC:       m.EnableGC,
		DebugGC:        m.DebugGC,
	}
}

// PrintGCStats 打印 GC 统计
func (p *Profiler) PrintGCStats() {
	stats := p.GetGCStats()

	fmt.Println("GC Statistics:")
	fmt.Printf("  LastGC:        %v\n", stats.LastGC)
	fmt.Printf("  NumGC:         %d\n", stats.NumGC)
	fmt.Printf("  PauseTotal:    %v\n", stats.PauseTotal)
	fmt.Printf("  NextGC:        %v MB\n", stats.NextGC/1024/1024)
	fmt.Printf("  NumForcedGC:   %d\n", stats.NumForcedGC)
	fmt.Printf("  GCCPUFraction: %.2f%%\n", stats.GCCPUFraction*100)
}

// RuntimeStats 运行时统计
type RuntimeStats struct {
	NumGoroutine int
	NumCPU       int
	NumCgoCall   int64
	GoVersion    string
	OS           string
	Arch         string
}

// GetRuntimeStats 获取运行时统计
func GetRuntimeStats() RuntimeStats {
	return RuntimeStats{
		NumGoroutine: runtime.NumGoroutine(),
		NumCPU:       runtime.NumCPU(),
		NumCgoCall:   runtime.NumCgoCall(),
		GoVersion:    runtime.Version(),
		OS:           runtime.GOOS,
		Arch:         runtime.GOARCH,
	}
}

// PrintRuntimeStats 打印运行时统计
func PrintRuntimeStats() {
	stats := GetRuntimeStats()

	fmt.Println("Runtime Statistics:")
	fmt.Printf("  Goroutines: %d\n", stats.NumGoroutine)
	fmt.Printf("  CPUs:       %d\n", stats.NumCPU)
	fmt.Printf("  CGO calls:  %d\n", stats.NumCgoCall)
	fmt.Printf("  Go version: %s\n", stats.GoVersion)
	fmt.Printf("  OS:         %s\n", stats.OS)
	fmt.Printf("  Arch:       %s\n", stats.Arch)
}

// Benchmark 基准测试辅助
type Benchmark struct {
	name   string
	start  time.Time
 laps   int
}

// NewBenchmark 创建基准测试
func NewBenchmark(name string) *Benchmark {
	return &Benchmark{
		name:  name,
		start: time.Now(),
		laps:  0,
	}
}

// Lap 记录一次循环
func (b *Benchmark) Lap() time.Duration {
	elapsed := time.Since(b.start)
	b.laps++
	b.start = time.Now()
	return elapsed
}

// End 结束基准测试
func (b *Benchmark) End() time.Duration {
	return time.Since(b.start)
}

// Report 报告结果
func (b *Benchmark) Report() {
	fmt.Printf("Benchmark '%s' completed %d laps\n", b.name, b.laps)
}

// CustomProfiler 自定义分析器
type CustomProfiler struct {
	mu       sync.Mutex
	metrics  map[string][]time.Duration
	counters map[string]int64
}

// NewCustomProfiler 创建自定义分析器
func NewCustomProfiler() *CustomProfiler {
	return &CustomProfiler{
		metrics:  make(map[string][]time.Duration),
		counters: make(map[string]int64),
	}
}

// RecordDuration 记录持续时间
func (cp *CustomProfiler) RecordDuration(name string, duration time.Duration) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.metrics[name] = append(cp.metrics[name], duration)
}

// Time 计时辅助
func (cp *CustomProfiler) Time(name string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		cp.RecordDuration(name, duration)
	}
}

// IncrementCounter 增加计数器
func (cp *CustomProfiler) IncrementCounter(name string, delta int64) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.counters[name] += delta
}

// GetMetrics 获取指标
func (cp *CustomProfiler) GetMetrics(name string) ([]time.Duration, bool) {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	metrics, exists := cp.metrics[name]
	return metrics, exists
}

// GetCounter 获取计数器
func (cp *CustomProfiler) GetCounter(name string) int64 {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	return cp.counters[name]
}

// Report 报告所有指标
func (cp *CustomProfiler) Report() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	fmt.Println("Custom Profiler Report:")
	fmt.Println("Durations:")
	for name, durations := range cp.metrics {
		if len(durations) == 0 {
			continue
		}

		total := time.Duration(0)
		min := durations[0]
		max := durations[0]

		for _, d := range durations {
			total += d
			if d < min {
				min = d
			}
			if d > max {
				max = d
			}
		}

		avg := total / time.Duration(len(durations))
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    Count: %d\n", len(durations))
		fmt.Printf("    Total: %v\n", total)
		fmt.Printf("    Avg:   %v\n", avg)
		fmt.Printf("    Min:   %v\n", min)
		fmt.Printf("    Max:   %v\n", max)
	}

	fmt.Println("Counters:")
	for name, count := range cp.counters {
		fmt.Printf("  %s: %d\n", name, count)
	}
}

// Reset 重置所有指标
func (cp *CustomProfiler) Reset() {
	cp.mu.Lock()
	defer cp.mu.Unlock()

	cp.metrics = make(map[string][]time.Duration)
	cp.counters = make(map[string]int64)
}
