// Package format 提供格式化功能
package format

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"time"
)

// Bytes 格式化字节数
func Bytes(bytes int64) string {
	const unit = 1024

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// BytesDecimal 格式化字节数（十进制）
func BytesDecimal(bytes int64) string {
	const unit = 1000

	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// Number 格式化数字（千分位）
func Number(n int64) string {
	str := fmt.Sprintf("%d", n)

	var result []byte
	for i := len(str) - 1; i >= 0; i-- {
		pos := len(str) - i - 1
		if pos > 0 && pos%3 == 0 {
			result = append([]byte{','}, result...)
		}
		result = append([]byte{str[i]}, result...)
	}

	return string(result)
}

// NumberFloat 格式化浮点数（千分位）
func NumberFloat(f float64, decimals int) string {
	str := fmt.Sprintf(fmt.Sprintf("%%.%df", decimals), f)

	parts := strings.Split(str, ".")
	integer := parts[0]

	var result []byte
	for i := len(integer) - 1; i >= 0; i-- {
		pos := len(integer) - i - 1
		if pos > 0 && pos%3 == 0 {
			result = append([]byte{','}, result...)
		}
		result = append([]byte{integer[i]}, result...)
	}

	if len(parts) > 1 {
		return string(result) + "." + parts[1]
	}

	return string(result)
}

// Percent 格式化百分比
func Percent(value, total int64) string {
	if total == 0 {
		return "0%"
	}
	percent := float64(value) / float64(total) * 100
	return fmt.Sprintf("%.2f%%", percent)
}

// PercentFloat 格式化百分比（浮点数）
func PercentFloat(value, total float64) string {
	if total == 0 {
		return "0%"
	}
	percent := value / total * 100
	return fmt.Sprintf("%.2f%%", percent)
}

// Duration 格式化时间间隔
func Duration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

// DurationPrecise 精确格式化时间间隔
func DurationPrecise(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%dμs", d.Microseconds())
	}
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

// Date 格式化日期
func Date(t time.Time) string {
	return t.Format("2006-01-02")
}

// DateTime 格式化日期时间
func DateTime(t time.Time) string {
	return t.Format("2006-01-02 15:04:05")
}

// Time 格式化时间
func Time(t time.Time) string {
	return t.Format("15:04:05")
}

// RelativeTime 相对时间
func RelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "刚刚"
	}
	if diff < time.Hour {
		minutes := int(diff.Minutes())
		return fmt.Sprintf("%d分钟前", minutes)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%d小时前", hours)
	}
	if diff < 30*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%d天前", days)
	}
	if diff < 365*24*time.Hour {
		months := int(diff.Hours() / 24 / 30)
		return fmt.Sprintf("%d个月前", months)
	}

	years := int(diff.Hours() / 24 / 365)
	return fmt.Sprintf("%d年前", years)
}

// Money 格式化金额
func Money(amount float64) string {
	return fmt.Sprintf("¥%.2f", amount)
}

// MoneyUS 美元格式
func MoneyUS(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

// MoneyEUR 欧元格式
func MoneyEUR(amount float64) string {
	return fmt.Sprintf("€%.2f", amount)
}

// Phone 格式化手机号（中国大陆）
func Phone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return fmt.Sprintf("%s %s %s", phone[:3], phone[3:7], phone[7:])
}

// IDCard 格式化身份证号（中国大陆）
func IDCard(id string) string {
	length := len(id)
	if length != 15 && length != 18 {
		return id
	}

	if length == 15 {
		return fmt.Sprintf("%s %s %s", id[:6], id[6:12], id[12:])
	}

	return fmt.Sprintf("%s %s %s %s", id[:6], id[6:10], id[10:14], id[14:])
}

// BankCard 格式化银行卡号
func BankCard(card string) string {
	length := len(card)
	if length < 13 {
		return card
	}

	var result []byte
	for i := 0; i < length; i++ {
		if i > 0 && i%4 == 0 {
			result = append(result, ' ')
		}
		result = append(result, card[i])
	}

	return string(result)
}

// MaskString 掩码字符串
func MaskString(s string, visible int) string {
	runes := []rune(s)
	length := len(runes)

	if length <= visible*2 {
		return strings.Repeat("*", length)
	}

	result := make([]rune, length)

	// 保留开头
	for i := 0; i < visible && i < length; i++ {
		result[i] = runes[i]
	}

	// 掩码中间
	for i := visible; i < length-visible; i++ {
		result[i] = '*'
	}

	// 保留结尾
	for i := length - visible; i < length; i++ {
		result[i] = runes[i]
	}

	return string(result)
}

// MaskPhone 掩码手机号
func MaskPhone(phone string) string {
	if len(phone) != 11 {
		return phone
	}
	return phone[:3] + "****" + phone[7:]
}

// MaskEmail 掩码邮箱
func MaskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return email
	}

	username := parts[0]
	if len(username) <= 2 {
		return email
	}

	maskedUsername := string([]rune(username)[0]) + "****" + string([]rune(username)[len(username)-1])
	return maskedUsername + "@" + parts[1]
}

// MaskIDCard 掩码身份证
func MaskIDCard(id string) string {
	length := len(id)
	if length < 8 {
		return id
	}

	visible := 4
	if length == 15 {
		visible = 3
	}

	return id[:visible] + strings.Repeat("*", length-visible*2) + id[length-visible:]
}

// Truncate 截断字符串
func Truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	return string(runes[:maxLen]) + "..."
}

// TruncateWithSuffix 截断字符串（自定义后缀）
func TruncateWithSuffix(s string, maxLen int, suffix string) string {
	runes := []rune(s)
	suffixLen := len([]rune(suffix))

	if len(runes) <= maxLen {
		return s
	}

	if maxLen <= suffixLen {
		maxLen = suffixLen + 1
	}

	return string(runes[:maxLen-suffixLen]) + suffix
}

// Ellipsis 省略字符串
func Ellipsis(s string, maxLen int) string {
	return Truncate(s, maxLen)
}

// Indent 缩进
func Indent(s string, indent string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// Dedent 去除缩进
func Dedent(s string) string {
	lines := strings.Split(s, "\n")

	// 找最小缩进
	minIndent := -1
	for _, line := range lines {
		if line == "" {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return s
	}

	// 去除缩进
	for i, line := range lines {
		if len(line) >= minIndent {
			lines[i] = line[minIndent:]
		}
	}

	return strings.Join(lines, "\n")
}

// Center 居中文本
func Center(s string, width int) string {
	length := len(s)
	if length >= width {
		return s
	}

	padding := width - length
	leftPadding := padding / 2
	rightPadding := padding - leftPadding

	return strings.Repeat(" ", leftPadding) + s + strings.Repeat(" ", rightPadding)
}

// LeftAlign 左对齐
func LeftAlign(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return s + strings.Repeat(" ", width-len(s))
}

// RightAlign 右对齐
func RightAlign(s string, width int) string {
	if len(s) >= width {
		return s
	}
	return strings.Repeat(" ", width-len(s)) + s
}

// PadLeft 左填充
func PadLeft(s, pad string, width int) string {
	for len(s) < width {
		s = pad + s
	}
	return s
}

// PadRight 右填充
func PadRight(s, pad string, width int) string {
	for len(s) < width {
		s = s + pad
	}
	return s
}

// PadBoth 两端填充
func PadBoth(s, pad string, width int) string {
	for len(s) < width {
		s = pad + s + pad
		// 如果填充后超出，去掉右边多余的
		if len(s) > width {
			s = s[:width]
		}
	}
	return s
}

// ToUpper 转大写
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower 转小写
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToTitle 转标题格式
func ToTitle(s string) string {
	return strings.ToTitle(s)
}

// Capitalize 首字母大写
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}

	runes := []rune(s)
	runes[0] = []rune(strings.ToUpper(string(runes[0])))[0]

	return string(runes)
}

// SnakeCase 转蛇形命名
func SnakeCase(s string) string {
	var result []rune

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '_')
		}
		result = append(result, r)
	}

	return strings.ToLower(string(result))
}

// KebabCase 转短横线命名
func KebabCase(s string) string {
	var result []rune

	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result = append(result, '-')
		}
		result = append(result, r)
	}

	return strings.ToLower(string(result))
}

// CamelCase 转驼峰命名
func CamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})

	if len(words) == 0 {
		return ""
	}

	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else {
			words[i] = strings.Title(strings.ToLower(word))
		}
	}

	return strings.Join(words, "")
}

// PascalCase 转帕斯卡命名
func PascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' ' || r == '.'
	})

	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}

	return strings.Join(words, "")
}

// Round 四舍五入
func Round(f float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Round(f*pow) / pow
}

// RoundUp 向上取整
func RoundUp(f float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Ceil(f*pow) / pow
}

// RoundDown 向下取整
func RoundDown(f float64, precision int) float64 {
	pow := math.Pow(10, float64(precision))
	return math.Floor(f*pow) / pow
}

// FormatFloat 格式化浮点数
func FormatFloat(f float64, precision int) string {
	return fmt.Sprintf(fmt.Sprintf("%%.%df", precision), f)
}

// FormatFloatWithThousand 格式化浮点数（千分位）
func FormatFloatWithThousand(f float64, precision int) string {
	str := fmt.Sprintf(fmt.Sprintf("%%.%df", precision), f)
	parts := strings.Split(str, ".")

	integer := parts[0]
	var result []byte

	for i := len(integer) - 1; i >= 0; i-- {
		pos := len(integer) - i - 1
		if pos > 0 && pos%3 == 0 {
			result = append([]byte{','}, result...)
		}
		result = append([]byte{integer[i]}, result...)
	}

	if len(parts) > 1 {
		return string(result) + "." + parts[1]
	}

	return string(result)
}

// Boolean 格式化布尔值
func Boolean(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// BooleanCN 格式化布尔值（中文）
func BooleanCN(b bool) string {
	if b {
		return "是"
	}
	return "否"
}

// BooleanYesNo 格式化布尔值（Yes/No）
func BooleanYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// Pluralize 复数化
func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, singular)
	}
	return fmt.Sprintf("%d %s", count, plural)
}

// Ordinal 序数词
func Ordinal(n int) string {
	suffix := "th"

	switch n % 10 {
	case 1:
		if n%100 != 11 {
			suffix = "st"
		}
	case 2:
		if n%100 != 12 {
			suffix = "nd"
		}
	case 3:
		if n%100 != 13 {
			suffix = "rd"
		}
	}

	return fmt.Sprintf("%d%s", n, suffix)
}

// Roman 罗马数字
func Roman(n int) string {
	if n <= 0 || n > 3999 {
		return ""
	}

	vals := []int{1000, 900, 500, 400, 100, 90, 50, 40, 10, 9, 5, 4, 1}
	syms := []string{"M", "CM", "D", "CD", "C", "XC", "L", "XL", "X", "IX", "V", "IV", "I"}

	var result bytes.Buffer

	for i := 0; i < len(vals); i++ {
		for n >= vals[i] {
			result.WriteString(syms[i])
			n -= vals[i]
		}
	}

	return result.String()
}
