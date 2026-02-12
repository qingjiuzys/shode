// Package progress 提供进度显示功能
package progress

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressBar 进度条
type ProgressBar struct {
	total     int64
	current   int64
	width     int
	prefix    string
	suffix    string
	theme     Theme
	mu        sync.Mutex
	startTime time.Time
}

// Theme 进度条主题
type Theme struct {
	BarStart   string
	BarEnd     string
	BarFill    string
	BarEmpty   string
	ShowPercent bool
}

// DefaultTheme 默认主题
var DefaultTheme = Theme{
	BarStart:    "[",
	BarEnd:      "]",
	BarFill:     "=",
	BarEmpty:    " ",
	ShowPercent: true,
}

// SimpleTheme 简单主题
var SimpleTheme = Theme{
	BarStart:    "",
	BarEnd:      "",
	BarFill:     "█",
	BarEmpty:    "░",
	ShowPercent: true,
}

// DotTheme 点主题
var DotTheme = Theme{
	BarStart:    "",
	BarEnd:      "",
	BarFill:     "•",
	BarEmpty:    "·",
	ShowPercent: false,
}

// NewProgressBar 创建进度条
func NewProgressBar(total int64) *ProgressBar {
	return &ProgressBar{
		total:     total,
		width:     50,
		theme:     DefaultTheme,
		startTime: time.Now(),
	}
}

// SetWidth 设置宽度
func (pb *ProgressBar) SetWidth(width int) *ProgressBar {
	pb.width = width
	return pb
}

// SetPrefix 设置前缀
func (pb *ProgressBar) SetPrefix(prefix string) *ProgressBar {
	pb.prefix = prefix
	return pb
}

// SetSuffix 设置后缀
func (pb *ProgressBar) SetSuffix(suffix string) *ProgressBar {
	pb.suffix = suffix
	return pb
}

// SetTheme 设置主题
func (pb *ProgressBar) SetTheme(theme Theme) *ProgressBar {
	pb.theme = theme
	return pb
}

// Add 增加进度
func (pb *ProgressBar) Add(delta int64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current += delta
	if pb.current > pb.total {
		pb.current = pb.total
	}
}

// Set 设置当前进度
func (pb *ProgressBar) Set(current int64) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current = current
	if pb.current > pb.total {
		pb.current = pb.total
	}
}

// Increment 自增
func (pb *ProgressBar) Increment() {
	pb.Add(1)
}

// Current 获取当前进度
func (pb *ProgressBar) Current() int64 {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	return pb.current
}

// Total 获取总数
func (pb *ProgressBar) Total() int64 {
	return pb.total
}

// Percent 获取百分比
func (pb *ProgressBar) Percent() float64 {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.total == 0 {
		return 0
	}
	return float64(pb.current) / float64(pb.total) * 100
}

// IsComplete 检查是否完成
func (pb *ProgressBar) IsComplete() bool {
	pb.mu.Lock()
	defer pb.mu.Unlock()
	return pb.current >= pb.total
}

// String 字符串表示
func (pb *ProgressBar) String() string {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	percent := float64(pb.current) / float64(pb.total)
	filledWidth := int(float64(pb.width) * percent)

	var bar strings.Builder

	// 前缀
	if pb.prefix != "" {
		bar.WriteString(pb.prefix)
		bar.WriteString(" ")
	}

	// 进度条开始
	bar.WriteString(pb.theme.BarStart)

	// 填充部分
	for i := 0; i < filledWidth; i++ {
		bar.WriteString(pb.theme.BarFill)
	}

	// 空白部分
	for i := filledWidth; i < pb.width; i++ {
		bar.WriteString(pb.theme.BarEmpty)
	}

	// 进度条结束
	bar.WriteString(pb.theme.BarEnd)

	// 百分比
	if pb.theme.ShowPercent {
		bar.WriteString(fmt.Sprintf(" %.1f%%", percent*100))
	}

	// 进度/总数
	bar.WriteString(fmt.Sprintf(" %d/%d", pb.current, pb.total))

	// 后缀
	if pb.suffix != "" {
		bar.WriteString(" ")
		bar.WriteString(pb.suffix)
	}

	return bar.String()
}

// Reset 重置进度条
func (pb *ProgressBar) Reset() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current = 0
	pb.startTime = time.Now()
}

// Elapsed 获取已用时间
func (pb *ProgressBar) Elapsed() time.Duration {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	return time.Since(pb.startTime)
}

// ETA 获取预计剩余时间
func (pb *ProgressBar) ETA() time.Duration {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.current == 0 {
		return 0
	}

	elapsed := time.Since(pb.startTime)
	remaining := pb.total - pb.current
	eta := time.Duration(int64(elapsed) * int64(remaining) / int64(pb.current))

	return eta
}

// Progress 进度管理器
type Progress struct {
	bars      []*ProgressBar
	mu        sync.Mutex
	autoPrint bool
}

// NewProgress 创建进度管理器
func NewProgress() *Progress {
	return &Progress{
		bars: make([]*ProgressBar, 0),
	}
}

// AddBar 添加进度条
func (p *Progress) AddBar(total int64) *ProgressBar {
	pb := NewProgressBar(total)

	p.mu.Lock()
	p.bars = append(p.bars, pb)
	p.mu.Unlock()

	return pb
}

// RemoveBar 移除进度条
func (p *Progress) RemoveBar(pb *ProgressBar) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, bar := range p.bars {
		if bar == pb {
			p.bars = append(p.bars[:i], p.bars[i+1:]...)
			break
		}
	}
}

// String 字符串表示
func (p *Progress) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()

	var lines []string
	for _, bar := range p.bars {
		lines = append(lines, bar.String())
	}

	return strings.Join(lines, "\n")
}

// Print 打印进度
func (p *Progress) Print() {
	fmt.Print("\r" + p.String())
}

// Println 打印进度（换行）
func (p *Progress) Println() {
	fmt.Println(p.String())
}

// Spinner 加载动画
type Spinner struct {
	frames   []string
	current  int
	active   bool
	mu       sync.Mutex
	prefix   string
	suffix   string
	stopChan chan struct{}
}

// DefaultFrames 默认动画帧
var DefaultFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

// NewSpinner 创建加载动画
func NewSpinner() *Spinner {
	return &Spinner{
		frames:   DefaultFrames,
		stopChan: make(chan struct{}),
	}
}

// SetFrames 设置动画帧
func (s *Spinner) SetFrames(frames []string) *Spinner {
	s.frames = frames
	return s
}

// SetPrefix 设置前缀
func (s *Spinner) SetPrefix(prefix string) *Spinner {
	s.prefix = prefix
	return s
}

// SetSuffix 设置后缀
func (s *Spinner) SetSuffix(suffix string) *Spinner {
	s.suffix = suffix
	return s
}

// Start 启动动画
func (s *Spinner) Start() {
	s.mu.Lock()
	if s.active {
		s.mu.Unlock()
		return
	}
	s.active = true
	s.mu.Unlock()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.mu.Lock()
				if !s.active {
					s.mu.Unlock()
					return
				}

				frame := s.frames[s.current%len(s.frames)]
				s.current++

				// 打印并清除
				fmt.Printf("\r%s%s%s", s.prefix, frame, s.suffix)
				s.mu.Unlock()
			case <-s.stopChan:
				return
			}
		}
	}()
}

// Stop 停止动画
func (s *Spinner) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active {
		return
	}

	s.active = false
	close(s.stopChan)
	s.stopChan = make(chan struct{})

	// 清除行
	fmt.Print("\r\033[K")
}

// IsActive 检查是否活动
func (s *Spinner) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// Counter 计数器
type Counter struct {
	value int64
	mu    sync.Mutex
}

// NewCounter 创建计数器
func NewCounter() *Counter {
	return &Counter{}
}

// Increment 自增
func (c *Counter) Increment() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
	return c.value
}

// Add 增加值
func (c *Counter) Add(delta int64) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += delta
	return c.value
}

// Get 获取值
func (c *Counter) Get() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Reset 重置
func (c *Counter) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value = 0
}

// Speedometer 速度计
type Speedometer struct {
	total    int64
	duration time.Duration
	mu       sync.Mutex
}

// NewSpeedometer 创建速度计
func NewSpeedometer() *Speedometer {
	return &Speedometer{}
}

// Add 添加进度
func (sm *Speedometer) Add(delta int64) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.total += delta
}

// Speed 获取速度（每秒）
func (sm *Speedometer) Speed() float64 {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.duration == 0 {
		return 0
	}

	return float64(sm.total) / sm.duration.Seconds()
}

// Reset 重置
func (sm *Speedometer) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.total = 0
	sm.duration = 0
}

// SetDuration 设置持续时间
func (sm *Speedometer) SetDuration(duration time.Duration) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.duration = duration
}

// FormatSpeed 格式化速度
func FormatSpeed(bytes float64) string {
	const unit = 1024

	if bytes < unit {
		return fmt.Sprintf("%.2f B/s", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %ciB/s", bytes/float64(div), "KMGTPE"[exp])
}
