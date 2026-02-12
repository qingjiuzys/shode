// Package timeutil 提供时间处理工具
package timeutil

import (
	"fmt"
	"time"
)

// Now 获取当前时间
func Now() time.Time {
	return time.Now()
}

// NowUnix 获取当前Unix时间戳（秒）
func NowUnix() int64 {
	return time.Now().Unix()
}

// NowUnixMilli 获取当前Unix时间戳（毫秒）
func NowUnixMilli() int64 {
	return time.Now().UnixMilli()
}

// NowUnixNano 获取当前Unix时间戳（纳秒）
func NowUnixNano() int64 {
	return time.Now().UnixNano()
}

// Format 格式化时间
func Format(t time.Time, layout string) string {
	return t.Format(layout)
}

// FormatWithDefault 格式化时间（默认格式）
func FormatWithDefault(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatDate 格式化日期
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatTime 格式化时间
func FormatTime(t time.Time) string {
	return t.Format("15:04:05")
}

// FormatDateTime 格式化日期时间
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatISO8601 格式化为ISO8601格式
func FormatISO8601(t time.Time) string {
	return t.Format(time.RFC3339)
}

// Parse 解析时间字符串
func Parse(s string, layout string) (time.Time, error) {
	return time.Parse(layout, s)
}

// ParseWithDefault 解析时间字符串（默认格式）
func ParseWithDefault(s string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", s)
}

// ParseISO8601 解析ISO8601格式
func ParseISO8601(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, s)
}

// MustParse 解析时间字符串（panic on error）
func MustParse(s string, layout string) time.Time {
	t, err := time.Parse(layout, s)
	if err != nil {
		panic(err)
	}
	return t
}

// Unix 转换Unix时间戳为时间
func Unix(sec int64) time.Time {
	return time.Unix(sec, 0)
}

// UnixMilli 转换毫秒时间戳为时间
func UnixMilli(msec int64) time.Time {
	return time.Unix(msec/1000, (msec%1000)*1e6)
}

// AddSeconds 增加秒数
func AddSeconds(t time.Time, seconds int64) time.Time {
	return t.Add(time.Duration(seconds) * time.Second)
}

// AddMinutes 增加分钟数
func AddMinutes(t time.Time, minutes int64) time.Time {
	return t.Add(time.Duration(minutes) * time.Minute)
}

// AddHours 增加小时数
func AddHours(t time.Time, hours int64) time.Time {
	return t.Add(time.Duration(hours) * time.Hour)
}

// AddDays 增加天数
func AddDays(t time.Time, days int64) time.Time {
	return t.AddDate(0, 0, int(days))
}

// AddWeeks 增加周数
func AddWeeks(t time.Time, weeks int64) time.Time {
	return t.AddDate(0, 0, int(weeks*7))
}

// AddMonths 增加月数
func AddMonths(t time.Time, months int64) time.Time {
	return t.AddDate(0, int(months), 0)
}

// AddYears 增加年数
func AddYears(t time.Time, years int64) time.Time {
	return t.AddDate(int(years), 0, 0)
}

// SubSeconds 减少秒数
func SubSeconds(t time.Time, seconds int64) time.Time {
	return t.Add(-time.Duration(seconds) * time.Second)
}

// SubMinutes 减少分钟数
func SubMinutes(t time.Time, minutes int64) time.Time {
	return t.Add(-time.Duration(minutes) * time.Minute)
}

// SubHours 减少小时数
func SubHours(t time.Time, hours int64) time.Time {
	return t.Add(-time.Duration(hours) * time.Hour)
}

// SubDays 减少天数
func SubDays(t time.Time, days int64) time.Time {
	return t.AddDate(0, 0, -int(days))
}

// Diff 计算时间差
func Diff(t1, t2 time.Time) time.Duration {
	return t1.Sub(t2)
}

// DiffSeconds 计算秒数差
func DiffSeconds(t1, t2 time.Time) int64 {
	return int64(t1.Sub(t2).Seconds())
}

// DiffMinutes 计算分钟数差
func DiffMinutes(t1, t2 time.Time) int64 {
	return int64(t1.Sub(t2).Minutes())
}

// DiffHours 计算小时数差
func DiffHours(t1, t2 time.Time) int64 {
	return int64(t1.Sub(t2).Hours())
}

// DiffDays 计算天数差
func DiffDays(t1, t2 time.Time) int64 {
	return int64(t1.Sub(t2).Hours() / 24)
}

// IsBefore 检查t1是否在t2之前
func IsBefore(t1, t2 time.Time) bool {
	return t1.Before(t2)
}

// IsAfter 检查t1是否在t2之后
func IsAfter(t1, t2 time.Time) bool {
	return t1.After(t2)
}

// IsEqual 检查t1和t2是否相等
func IsEqual(t1, t2 time.Time) bool {
	return t1.Equal(t2)
}

// IsBetween 检查时间是否在范围内
func IsBetween(t, start, end time.Time) bool {
	return (t.Equal(start) || t.After(start)) && (t.Equal(end) || t.Before(end))
}

// IsToday 检查是否为今天
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// IsYesterday 检查是否为昨天
func IsYesterday(t time.Time) bool {
	yesterday := time.Now().AddDate(0, 0, -1)
	return t.Year() == yesterday.Year() && t.Month() == yesterday.Month() && t.Day() == yesterday.Day()
}

// IsTomorrow 检查是否为明天
func IsTomorrow(t time.Time) bool {
	tomorrow := time.Now().AddDate(0, 0, 1)
	return t.Year() == tomorrow.Year() && t.Month() == tomorrow.Month() && t.Day() == tomorrow.Day()
}

// IsThisWeek 检查是否本周
func IsThisWeek(t time.Time) bool {
	now := time.Now()
	year, week := now.ISOWeek()
	tYear, tWeek := t.ISOWeek()
	return year == tYear && week == tWeek
}

// IsThisMonth 检查是否本月
func IsThisMonth(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month()
}

// IsThisYear 检查是否本年
func IsThisYear(t time.Time) bool {
	return t.Year() == time.Now().Year()
}

// StartOfDay 获取一天的开始时间
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取一天的结束时间
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// StartOfWeek 获取一周的开始时间（周一）
func StartOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	return t.AddDate(0, 0, -int(weekday)+1)
}

// EndOfWeek 获取一周的结束时间（周日）
func EndOfWeek(t time.Time) time.Time {
	return StartOfWeek(t).AddDate(0, 0, 7)
}

// StartOfMonth 获取月的开始时间
func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 获取月的结束时间
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, -1)
}

// StartOfYear 获取年的开始时间
func StartOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear 获取年的结束时间
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// StartOfHour 获取小时的开始时间
func StartOfHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

// EndOfHour 获取小时的结束时间
func EndOfHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// StartOfMinute 获取分钟的开始时间
func StartOfMinute(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 0, 0, t.Location())
}

// EndOfMinute 获取分钟的结束时间
func EndOfMinute(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), 59, int(time.Second-time.Nanosecond), t.Location())
}

// BeginOfDay StartOfDay的别名
func BeginOfDay(t time.Time) time.Time {
	return StartOfDay(t)
}

// BeginOfWeek StartOfWeek的别名
func BeginOfWeek(t time.Time) time.Time {
	return StartOfWeek(t)
}

// BeginOfMonth StartOfMonth的别名
func BeginOfMonth(t time.Time) time.Time {
	return StartOfMonth(t)
}

// BeginOfYear StartOfYear的别名
func BeginOfYear(t time.Time) time.Time {
	return StartOfYear(t)
}

// Age 计算年龄
func Age(birthday time.Time) int {
	now := time.Now()
	years := now.Year() - birthday.Year()

	// 如果还没到生日，减1岁
	if now.Month() < birthday.Month() || (now.Month() == birthday.Month() && now.Day() < birthday.Day()) {
		years--
	}

	return years
}

// DaysInMonth 获取月份的天数
func DaysInMonth(t time.Time) int {
	return EndOfMonth(t).Day()
}

// IsLeapYear 检查是否为闰年
func IsLeapYear(year int) bool {
	if year%4 != 0 {
		return false
	} else if year%100 != 0 {
		return true
	} else {
		return year%400 == 0
	}
}

// WeekOfYear 获取周数
func WeekOfYear(t time.Time) (year, week int) {
	return t.ISOWeek()
}

// DayOfWeek 获取星期几（0=周日, 1=周一, ..., 6=周六）
func DayOfWeek(t time.Time) time.Weekday {
	return t.Weekday()
}

// Quarter 获取季度
func Quarter(t time.Time) int {
	month := t.Month()
	if month <= 3 {
		return 1
	} else if month <= 6 {
		return 2
	} else if month <= 9 {
		return 3
	}
	return 4
}

// StartOfQuarter 获取季度的开始时间
func StartOfQuarter(t time.Time) time.Time {
	quarter := Quarter(t)
	month := time.Month((quarter-1)*3 + 1)
	return time.Date(t.Year(), month, 1, 0, 0, 0, 0, t.Location())
}

// EndOfQuarter 获取季度的结束时间
func EndOfQuarter(t time.Time) time.Time {
	return StartOfQuarter(t).AddDate(0, 3, -1)
}

// Truncate 截断时间到指定精度
func Truncate(t time.Time, precision time.Duration) time.Time {
	return t.Truncate(precision)
}

// Round 舍入时间到指定精度
func Round(t time.Time, precision time.Duration) time.Time {
	if precision <= 0 {
		return t
	}

	// 计算余数
	remainder := t.Sub(time.Time{}) % precision
	if remainder < 0 {
		remainder += precision
	}

	// 如果余数大于精度的一半，向上舍入
	if remainder*2 > precision {
		return t.Add(precision - remainder)
	}
	return t.Add(-remainder)
}

// Range 时间范围
type Range struct {
	Start time.Time
	End   time.Time
}

// NewRange 创建时间范围
func NewRange(start, end time.Time) *Range {
	return &Range{
		Start: start,
		End:   end,
	}
}

// Contains 检查时间是否在范围内
func (r *Range) Contains(t time.Time) bool {
	return (t.Equal(r.Start) || t.After(r.Start)) && (t.Equal(r.End) || t.Before(r.End))
}

// Duration 获取持续时间
func (r *Range) Duration() time.Duration {
	return r.End.Sub(r.Start)
}

// Overlaps 检查范围是否重叠
func (r *Range) Overlaps(other *Range) bool {
	return r.Start.Before(other.End) && r.End.After(other.Start)
}

// String 字符串表示
func (r *Range) String() string {
	return fmt.Sprintf("[%s, %s]", r.Start.Format(time.RFC3339), r.End.Format(time.RFC3339))
}

// Since 计算从某个时间到现在的时间差
func Since(t time.Time) time.Duration {
	return time.Since(t)
}

// Until 计算从现在到某个时间的时间差
func Until(t time.Time) time.Duration {
	return time.Until(t)
}

// Sleep 睡眠指定时间
func Sleep(duration time.Duration) {
	time.Sleep(duration)
}

// After 在指定时间后执行
func After(duration time.Duration) <-chan time.Time {
	return time.After(duration)
}

// Tick 每隔一段时间执行
func Tick(duration time.Duration) <-chan time.Time {
	return time.Tick(duration)
}

// Timer 创建定时器
func Timer(duration time.Duration) *time.Timer {
	return time.NewTimer(duration)
}

// Ticker 创建周期性定时器
func Ticker(duration time.Duration) *time.Ticker {
	return time.NewTicker(duration)
}

// Measure 测量函数执行时间
func Measure(fn func()) time.Duration {
	start := time.Now()
	fn()
	return time.Since(start)
}

// MeasureWithResult 测量函数执行时间并返回结果
func MeasureWithResult[T any](fn func() T) (T, time.Duration) {
	start := time.Now()
	result := fn()
	duration := time.Since(start)
	return result, duration
}

// RetryUntil 重试直到成功或超时
func RetryUntil(fn func() error, maxAttempts int, interval time.Duration) error {
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err == nil {
			return nil
		} else {
			lastErr = err
			if i < maxAttempts-1 {
				time.Sleep(interval)
			}
		}
	}
	return lastErr
}

// Timeout 执行函数带超时
func Timeout(fn func() error, timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- fn()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("timeout after %v", timeout)
	}
}

// TimeTracker 时间追踪器
type TimeTracker struct {
	start time.Time
 laps []lap
}

type lap struct {
	name string
	time time.Time
}

// NewTimeTracker 创建时间追踪器
func NewTimeTracker() *TimeTracker {
	return &TimeTracker{
		start: time.Now(),
	 laps:  make([]lap, 0),
	}
}

// Start 开始计时
func (tt *TimeTracker) Start() {
	tt.start = time.Now()
	tt.laps = make([]lap, 0)
}

// Lap 记录一个时间点
func (tt *TimeTracker) Lap(name string) {
	tt.laps = append(tt.laps, lap{
		name: name,
		time: time.Now(),
	})
}

// Elapsed 获取总耗时
func (tt *TimeTracker) Elapsed() time.Duration {
	return time.Since(tt.start)
}

// LapDuration 获取特定圈的时间差
func (tt *TimeTracker) LapDuration(name string) time.Duration {
	for i, lap := range tt.laps {
		if lap.name == name {
			if i == 0 {
				return lap.time.Sub(tt.start)
			}
			return lap.time.Sub(tt.laps[i-1].time)
		}
	}
	return 0
}

// Laps 获取所有圈的信息
func (tt *TimeTracker) Laps() []lap {
	return tt.laps
}

// Reset 重置追踪器
func (tt *TimeTracker) Reset() {
	tt.start = time.Now()
	tt.laps = make([]lap, 0)
}
