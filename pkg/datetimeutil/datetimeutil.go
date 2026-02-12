// Package datetimeutil 提供日期时间处理工具
package datetimeutil

import (
	"fmt"
	"math"
	"time"
)

// Now 获取当前时间
func Now() time.Time {
	return time.Now()
}

// Today 获取今天的日期（时间部分为0）
func Today() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

// Tomorrow 获取明天的日期
func Tomorrow() time.Time {
	return Today().AddDate(0, 0, 1)
}

// Yesterday 获取昨天的日期
func Yesterday() time.Time {
	return Today().AddDate(0, 0, -1)
}

// BeginningOfDay 获取一天的开始
func BeginningOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay 获取一天的结束
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// BeginningOfWeek 获取一周的开始
func BeginningOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -weekday+1)
}

// EndOfWeek 获取一周的结束
func EndOfWeek(t time.Time) time.Time {
	return BeginningOfWeek(t).AddDate(0, 0, 7).Add(-time.Nanosecond)
}

// BeginningOfMonth 获取月的开始
func BeginningOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth 获取月的结束
func EndOfMonth(t time.Time) time.Time {
	return BeginningOfMonth(t).AddDate(0, 1, 0).Add(-time.Nanosecond)
}

// BeginningOfYear 获取年的开始
func BeginningOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 1, 1, 0, 0, 0, 0, t.Location())
}

// EndOfYear 获取年的结束
func EndOfYear(t time.Time) time.Time {
	return time.Date(t.Year(), 12, 31, 23, 59, 59, 999999999, t.Location())
}

// AddDays 添加天数
func AddDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, days)
}

// AddWeeks 添加周数
func AddWeeks(t time.Time, weeks int) time.Time {
	return t.AddDate(0, 0, weeks*7)
}

// AddMonths 添加月数
func AddMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, months, 0)
}

// AddYears 添加年数
func AddYears(t time.Time, years int) time.Time {
	return t.AddDate(years, 0, 0)
}

// SubDays 减去天数
func SubDays(t time.Time, days int) time.Time {
	return t.AddDate(0, 0, -days)
}

// SubWeeks 减去周数
func SubWeeks(t time.Time, weeks int) time.Time {
	return t.AddDate(0, 0, -weeks*7)
}

// SubMonths 减去月数
func SubMonths(t time.Time, months int) time.Time {
	return t.AddDate(0, -months, 0)
}

// SubYears 减去年数
func SubYears(t time.Time, years int) time.Time {
	return t.AddDate(-years, 0, 0)
}

// DiffInDays 日期相差天数
func DiffInDays(a, b time.Time) int {
	duration := a.Sub(b)
	return int(math.Abs(duration.Hours() / 24))
}

// DiffInHours 相差小时数
func DiffInHours(a, b time.Time) int {
	duration := a.Sub(b)
	return int(math.Abs(duration.Hours()))
}

// DiffInMinutes 相差分钟数
func DiffInMinutes(a, b time.Time) int {
	duration := a.Sub(b)
	return int(math.Abs(duration.Minutes()))
}

// DiffInSeconds 相差秒数
func DiffInSeconds(a, b time.Time) int {
	duration := a.Sub(b)
	return int(math.Abs(duration.Seconds()))
}

// Age 计算年龄
func Age(birthday time.Time) int {
	now := time.Now()
	years := now.Year() - birthday.Year()

	// 如果还没到生日，减1岁
	if now.Month() < birthday.Month() ||
		(now.Month() == birthday.Month() && now.Day() < birthday.Day()) {
		years--
	}

	return years
}

// IsBefore 是否在之前
func IsBefore(a, b time.Time) bool {
	return a.Before(b)
}

// IsAfter 是否在之后
func IsAfter(a, b time.Time) bool {
	return a.After(b)
}

// IsBetween 是否在之间
func IsBetween(t, start, end time.Time) bool {
	return t.After(start) && t.Before(end)
}

// IsToday 是否是今天
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// IsTomorrow 是否是明天
func IsTomorrow(t time.Time) bool {
	return IsToday(t.AddDate(0, 0, -1))
}

// IsYesterday 是否是昨天
func IsYesterday(t time.Time) bool {
	return IsToday(t.AddDate(0, 0, 1))
}

// IsWeekend 是否是周末
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWeekday 是否是工作日
func IsWeekday(t time.Time) bool {
	return !IsWeekend(t)
}

// IsLeapYear 是否是闰年
func IsLeapYear(year int) bool {
	if year%4 != 0 {
		return false
	} else if year%100 != 0 {
		return true
	} else {
		return year%400 == 0
	}
}

// DaysInMonth 获取月的天数
func DaysInMonth(year int, month time.Month) int {
	if month == time.February {
		if IsLeapYear(year) {
			return 29
		}
		return 28
	}

	switch month {
	case time.April, time.June, time.September, time.November:
		return 30
	default:
		return 31
	}
}

// FirstDayOfMonth 获取月的第一天
func FirstDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// LastDayOfMonth 获取月的最后一天
func LastDayOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), DaysInMonth(t.Year(), t.Month()), 0, 0, 0, 0, t.Location())
}

// Quarter 获取季度
func Quarter(t time.Time) int {
	month := int(t.Month())
	return (month-1)/3 + 1
}

// BeginningOfQuarter 获取季度的开始
func BeginningOfQuarter(t time.Time) time.Time {
	month := t.Month()
	quarter := (month-1)/3

	firstMonth := time.Month(quarter*3 + 1)
	return time.Date(t.Year(), firstMonth, 1, 0, 0, 0, 0, t.Location())
}

// EndOfQuarter 获取季度的结束
func EndOfQuarter(t time.Time) time.Time {
	return BeginningOfQuarter(t).AddDate(0, 3, 0).Add(-time.Nanosecond)
}

// WeekOfYear 获取周数
func WeekOfYear(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

// DayOfYear 获取一年中的第几天
func DayOfYear(t time.Time) int {
	beginningOfYear := BeginningOfYear(t)
	return int(t.Sub(beginningOfYear).Hours()/24) + 1
}

// Parse 解析日期字符串
func Parse(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// ParseInLocation 在指定时区解析日期
func ParseInLocation(layout, value string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(layout, value, loc)
}

// ParseISO 解析ISO格式日期
func ParseISO(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

// ParseDate 解析日期（YYYY-MM-DD）
func ParseDate(value string) (time.Time, error) {
	return time.Parse("2006-01-02", value)
}

// ParseDateTime 解析日期时间（YYYY-MM-DD HH:MM:SS）
func ParseDateTime(value string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", value)
}

// Format 格式化日期
func Format(t time.Time, layout string) string {
	return t.Format(layout)
}

// FormatISO 格式化为ISO格式
func FormatISO(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FormatDate 格式化为日期
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// FormatDateTime 格式化为日期时间
func FormatDateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// FormatTime 格式化为时间
func FormatTime(t time.Time) string {
	return t.Format("15:04:05")
}

// FormatFriendly 友好的时间格式
func FormatFriendly(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d分钟前", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1小时前"
		}
		return fmt.Sprintf("%d小时前", hours)
	} else if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "昨天"
		}
		return fmt.Sprintf("%d天前", days)
	} else if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / 24 / 7)
		if weeks == 1 {
			return "1周前"
		}
		return fmt.Sprintf("%d周前", weeks)
	} else if diff < 365*24*time.Hour {
		months := int(diff.Hours() / 24 / 30)
		if months == 1 {
			return "1个月前"
		}
		return fmt.Sprintf("%d个月前", months)
	} else {
		years := int(diff.Hours() / 24 / 365)
		if years == 1 {
			return "1年前"
		}
		return fmt.Sprintf("%d年前", years)
	}
}

// FormatRelative 相对时间格式
func FormatRelative(t time.Time) string {
	now := time.Now()
	if t.After(now) {
		diff := t.Sub(now)
		if diff < time.Minute {
			return "片刻之后"
		} else if diff < time.Hour {
			return fmt.Sprintf("%d分钟后", int(diff.Minutes()))
		} else if diff < 24*time.Hour {
			hours := int(diff.Hours())
			if hours == 1 {
				return "1小时后"
			}
			return fmt.Sprintf("%d小时后", hours)
		} else if diff < 7*24*time.Hour {
			days := int(diff.Hours() / 24)
			if days == 1 {
				return "明天"
			}
			return fmt.Sprintf("%d天后", days)
		}
	}

	return FormatFriendly(t)
}

// Timezone 获取时区
func Timezone(name string) (*time.Location, error) {
	return time.LoadLocation(name)
}

// InTimezone 转换时区
func InTimezone(t time.Time, name string) (time.Time, error) {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.Time{}, err
	}
	return t.In(loc), nil
}

// UTC 转换为UTC时间
func UTC(t time.Time) time.Time {
	return t.UTC()
}

// Local 转换为本地时间
func Local(t time.Time) time.Time {
	return t.Local()
}

// Unix 转换为Unix时间戳
func Unix(t time.Time) int64 {
	return t.Unix()
}

// UnixMilli 转换为Unix毫秒时间戳
func UnixMilli(t time.Time) int64 {
	return t.UnixMilli()
}

// FromUnix 从Unix时间戳转换
func FromUnix(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// FromUnixMilli 从Unix毫秒时间戳转换
func FromUnixMilli(timestamp int64) time.Time {
	return time.UnixMilli(timestamp)
}

// StartOfWeek 获取本周开始
func StartOfWeek(t time.Time) time.Time {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	return t.AddDate(0, 0, -(weekday-1))
}

// RangeOfWeek 获取本周的日期范围
func RangeOfWeek(t time.Time) (time.Time, time.Time) {
	return StartOfWeek(t), EndOfWeek(t)
}

// RangeOfMonth 获取本月的日期范围
func RangeOfMonth(t time.Time) (time.Time, time.Time) {
	return FirstDayOfMonth(t), LastDayOfMonth(t)
}

// RangeOfQuarter 获取本季度的日期范围
func RangeOfQuarter(t time.Time) (time.Time, time.Time) {
	return BeginningOfQuarter(t), EndOfQuarter(t)
}

// RangeOfYear 获取本年的日期范围
func RangeOfYear(t time.Time) (time.Time, time.Time) {
	return BeginningOfYear(t), EndOfYear(t)
}

// IsValid 验证日期是否有效
func IsValid(year int, month time.Month, day int) bool {
	if year < 1 || month < 1 || month > 12 || day < 1 {
		return false
	}

	daysInMonth := DaysInMonth(year, month)
	return day <= daysInMonth
}

// Date 创建日期
func Date(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// DateTime 创建日期时间
func DateTime(year int, month time.Month, day, hour, minute, second int) time.Time {
	return time.Date(year, month, day, hour, minute, second, 0, time.UTC)
}

// Midnight 获取午夜时间
func Midnight(t time.Time) time.Time {
	return BeginningOfDay(t)
}

// Noon 获取正午时间
func Noon(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 12, 0, 0, 0, t.Location())
}

// DurationString 格式化时间段
func DurationString(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	} else if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	} else if d < 24*time.Hour {
		return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
	} else {
		days := int(d.Hours() / 24)
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%dd%dh", days, hours)
	}
}

// Milliseconds 获取毫秒数
func Milliseconds(t time.Time) int64 {
	return t.UnixMilli()
}

// Seconds 获取秒数
func Seconds(t time.Time) int64 {
	return t.Unix()
}

// Truncate 截断到指定精度
func Truncate(t time.Time, precision time.Duration) time.Time {
	return t.Truncate(precision)
}

// Round 四舍五入到指定精度
func Round(t time.Time, precision time.Duration) time.Time {
	return t.Round(precision)
}
