// Package profiler 提供性能分析功能
package profiler

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
	cpuProfile     *os.File
	memProfile     *os.File
	blockProfile   *os.File
	mutexProfile   *os.File
	enabled        bool
	startTime      time.Time
	mu             sync.Mutex
	tickers        []*time.Ticker
	recordMemStats bool
}

// Config 分析器配置
type Config struct {
	CPUProfile     string
	MemProfile     string
	BlockProfile   string
	MutexProfile   string
	RecordMemStats bool
}

// NewProfiler 创建分析器
func NewProfiler(config *Config) *Profiler {
	p := &Profiler{
		enabled:        true,
		startTime:      time.Now(),
		recordMemStats: config.RecordMemStats,
	}

	// CPU 性能分析
	if config.CPUProfile != "" {
		f, err := os.Create(config.CPUProfile)
		if err != nil {
			fmt.Printf("Failed to create CPU profile: %v\n", err)
		} else {
			p.cpuProfile = f
			if err := pprof.StartCPUProfile(f); err != nil {
				fmt.Printf("Failed to start CPU profile: %v\n", err)
				f.Close()
				p.cpuProfile = nil
			}
		}
	}

	// Block 性能分析
	if config.BlockProfile != "" {
		runtime.SetBlockProfileRate(1)
		f, err := os.Create(config.BlockProfile)
		if err != nil {
			fmt.Printf("Failed to create block profile: %v\n", err)
		} else {
			p.blockProfile = f
		}
	}

	// Mutex 性能分析
	if config.MutexProfile != "" {
		runtime.SetMutexProfileFraction(1)
		f, err := os.Create(config.MutexProfile)
		if err != nil {
			fmt.Printf("Failed to create mutex profile: %v\n", err)
		} else {
			p.mutexProfile = f
		}
	}

	return p
}

// Stop 停止分析
func (p *Profiler) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.enabled {
		return
	}

	// 停止 CPU 分析
	if p.cpuProfile != nil {
		pprof.StopCPUProfile()
		p.cpuProfile.Close()
		p.cpuProfile = nil
	}

	// 写入内存分析
	if p.memProfile != nil {
		if err := pprof.WriteHeapProfile(p.memProfile); err != nil {
			fmt.Printf("Failed to write memory profile: %v\n", err)
		}
		p.memProfile.Close()
		p.memProfile = nil
	}

	// 写入 Block 分析
	if p.blockProfile != nil {
		if err := pprof.Lookup("block").WriteTo(p.blockProfile, 0); err != nil {
			fmt.Printf("Failed to write block profile: %v\n", err)
		}
		p.blockProfile.Close()
		p.blockProfile = nil
	}

	// 写入 Mutex 分析
	if p.mutexProfile != nil {
		if err := pprof.Lookup("mutex").WriteTo(p.mutexProfile, 0); err != nil {
			fmt.Printf("Failed to write mutex profile: %v\n", err)
		}
		p.mutexProfile.Close()
		p.mutexProfile = nil
	}

	// 停止所有定时器
	for _, ticker := range p.tickers {
		ticker.Stop()
	}
	p.tickers = nil

	p.enabled = false
	fmt.Printf("Profiler stopped. Duration: %v\n", time.Since(p.startTime))
}

// Snapshot 获取内存快照
func (p *Profiler) Snapshot(filename string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return pprof.WriteHeapProfile(f)
}

// GetMemStats 获取内存统计
func (p *Profiler) GetMemStats() *runtime.MemStats {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return &stats
}

// PrintMemStats 打印内存统计
func (p *Profiler) PrintMemStats() {
	stats := p.GetMemStats()

	fmt.Println("\n=== Memory Statistics ===")
	fmt.Printf("Alloc:        %v MB\n", stats.Alloc/1024/1024)
	fmt.Printf("TotalAlloc:   %v MB\n", stats.TotalAlloc/1024/1024)
	fmt.Printf("Sys:          %v MB\n", stats.Sys/1024/1024)
	fmt.Printf("HeapAlloc:    %v MB\n", stats.HeapAlloc/1024/1024)
	fmt.Printf("HeapSys:      %v MB\n", stats.HeapSys/1024/1024)
	fmt.Printf("HeapObjects:  %d\n", stats.HeapObjects)
	fmt.Printf("StackInuse:   %v MB\n", stats.StackInuse/1024/1024)
	fmt.Printf("NextGC:       %v MB\n", stats.NextGC/1024/1024)
	fmt.Printf("NumGC:        %d\n", stats.NumGC)
	fmt.Printf("Goroutines:   %d\n", runtime.NumGoroutine())
	fmt.Println("========================\n")
}

// StartMemStatsMonitor 启动内存监控
func (p *Profiler) StartMemStatsMonitor(interval time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()

	ticker := time.NewTicker(interval)
	p.tickers = append(p.tickers, ticker)

	go func() {
		for range ticker.C {
			p.PrintMemStats()
		}
	}()

	fmt.Printf("Started memory stats monitor (interval: %v)\n", interval)
}

// Benchmark 基准测试辅助
type Benchmark struct {
	name string
}

// NewBenchmark 创建基准测试
func NewBenchmark(name string) *Benchmark {
	return &Benchmark{name: name}
}

// Run 运行基准测试
func (b *Benchmark) Run(fn func()) time.Duration {
	runtime.GC() // 清理内存
	start := time.Now()
	fn()
	duration := time.Since(start)
	fmt.Printf("%s: %v\n", b.name, duration)
	return duration
}

// RunMultiple 运行多次基准测试
func (b *Benchmark) RunMultiple(iterations int, fn func()) time.Duration {
	var total time.Duration
	for i := 0; i < iterations; i++ {
		runtime.GC()
		start := time.Now()
		fn()
		total += time.Since(start)
	}

	avg := total / time.Duration(iterations)
	fmt.Printf("%s (avg of %d runs): %v\n", b.name, iterations, avg)
	return avg
}

// Comparison 比较两个函数的性能
func Comparison(name1, name2 string, fn1, fn2 func()) time.Duration {
	b1 := NewBenchmark(name1)
	b2 := NewBenchmark(name2)

	d1 := b1.Run(fn1)
	d2 := b2.Run(fn2)

	diff := d1 - d2
	if diff > 0 {
		fmt.Printf("%s is faster by %v\n", name2, diff)
	} else {
		fmt.Printf("%s is faster by %v\n", name1, -diff)
	}

	return diff
}

// TraceGoroutine 追踪 Goroutine
func TraceGoroutine(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return pprof.Lookup("goroutine").WriteTo(f, 2)
}

// TraceThread 追踪线程
func TraceThread(filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return pprof.Lookup("threadcreate").WriteTo(f, 2)
}
