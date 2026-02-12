// Package numutil 提供数字处理工具
package numutil

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
)

// Abs 绝对值
func Abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// AbsFloat 浮点数绝对值
func AbsFloat(n float64) float64 {
	if n < 0 {
		return -n
	}
	return n
}

// Max 取最大值
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Min 取最小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// MaxFloat 浮点数最大值
func MaxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// MinFloat 浮点数最小值
func MinFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// Clamp 限制范围
func Clamp(n, min, max int) int {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// ClampFloat 浮点数限制范围
func ClampFloat(n, min, max float64) float64 {
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

// Range 生成范围
func Range(min, max int) []int {
	result := make([]int, max-min)
	for i := range result {
		result[i] = min + i
	}
	return result
}

// Sum 求和
func Sum(nums []int) int {
	sum := 0
	for _, n := range nums {
		sum += n
	}
	return sum
}

// SumFloat 浮点数求和
func SumFloat(nums []float64) float64 {
	sum := 0.0
	for _, n := range nums {
		sum += n
	}
	return sum
}

// Average 平均值
func Average(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}
	return float64(Sum(nums)) / float64(len(nums))
}

// AverageFloat 浮点数平均值
func AverageFloat(nums []float64) float64 {
	if len(nums) == 0 {
		return 0
	}
	return SumFloat(nums) / float64(len(nums))
}

// Product 乘积
func Product(nums []int) int {
	product := 1
	for _, n := range nums {
		product *= n
	}
	return product
}

// ProductFloat 浮点数乘积
func ProductFloat(nums []float64) float64 {
	product := 1.0
	for _, n := range nums {
		product *= n
	}
	return product
}

// Round 四舍五入
func Round(n float64) int {
	return int(n + 0.5)
}

// RoundFloat 四舍五入到指定位数
func RoundFloat(n float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(n*factor) / factor
}

// Floor 向下取整
func Floor(n float64) int {
	return int(math.Floor(n))
}

// Ceil 向上取整
func Ceil(n float64) int {
	return int(math.Ceil(n))
}

// IsEven 检查是否为偶数
func IsEven(n int) bool {
	return n%2 == 0
}

// IsOdd 检查是否为奇数
func IsOdd(n int) bool {
	return n%2 != 0
}

// IsPositive 检查是否为正数
func IsPositive(n int) bool {
	return n > 0
}

// IsNegative 检查是否为负数
func IsNegative(n int) bool {
	return n < 0
}

// IsZero 检查是否为零
func IsZero(n int) bool {
	return n == 0
}

// GCD 最大公约数
func GCD(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// LCM 最小公倍数
func LCM(a, b int) int {
	return a * b / GCD(a, b)
}

// Pow 幂运算
func Pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}

// PowFloat 浮点数幂运算
func PowFloat(base, exp float64) float64 {
	return math.Pow(base, exp)
}

// Sqrt 平方根
func Sqrt(n float64) float64 {
	return math.Sqrt(n)
}

// Log 对数
func Log(n, base float64) float64 {
	return math.Log(n) / math.Log(base)
}

// Log10 以10为底的对数
func Log10(n float64) float64 {
	return math.Log10(n)
}

// Ln 自然对数
func Ln(n float64) float64 {
	return math.Log(n)
}

// Sin 正弦
func Sin(n float64) float64 {
	return math.Sin(n)
}

// Cos 余弦
func Cos(n float64) float64 {
	return math.Cos(n)
}

// Tan 正切
func Tan(n float64) float64 {
	return math.Tan(n)
}

// DegreesToDegrees 弧度转角度
func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// RadiansToDegrees 角度转弧度
func RadiansToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// Percent 百分比
func Percent(value, total float64) float64 {
	if total == 0 {
		return 0
	}
	return (value / total) * 100
}

// PercentageChange 百分比变化
func PercentageChange(old, new float64) float64 {
	if old == 0 {
		return 0
	}
	return ((new - old) / old) * 100
}

// IsPrime 检查是否为质数
func IsPrime(n int) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}

	for i := 5; i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}

	return true
}

// Fibonacci 斐波那契数列
func Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	a, b := 0, 1
	for i := 2; i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// Factorial 阶乘
func Factorial(n int) int {
	if n <= 1 {
		return 1
	}
	result := 1
	for i := 2; i <= n; i++ {
		result *= i
	}
	return result
}

// FactorialFloat 浮点数阶乘
func FactorialFloat(n float64) float64 {
	result := 1.0
	for i := 2.0; i <= n; i++ {
		result *= i
	}
	return result
}

// IsBetween 检查是否在范围内
func IsBetween(n, min, max int) bool {
	return n >= min && n <= max
}

// IsBetweenFloat 浮点数检查是否在范围内
func IsBetweenFloat(n, min, max float64) bool {
	return n >= min && n <= max
}

// InRange 检查值是否在切片中
func InRange(n int, nums []int) bool {
	for _, num := range nums {
		if num == n {
			return true
		}
	}
	return false
}

// IndexOf 查找索引
func IndexOf(n int, nums []int) int {
	for i, num := range nums {
		if num == n {
			return i
		}
	}
	return -1
}

// Contains 检查是否包含
func Contains(nums []int, n int) bool {
	return IndexOf(n, nums) != -1
}

// Count 统计出现次数
func Count(nums []int, n int) int {
	count := 0
	for _, num := range nums {
		if num == n {
			count++
		}
	}
	return count
}

// Unique 去重
func Unique(nums []int) []int {
	seen := make(map[int]bool)
	result := []int{}

	for _, num := range nums {
		if !seen[num] {
			seen[num] = true
			result = append(result, num)
		}
	}

	return result
}

// Reverse 反转
func Reverse(nums []int) []int {
	result := make([]int, len(nums))
	copy(result, nums)

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// Sort 排序
func Sort(nums []int) []int {
	result := make([]int, len(nums))
	copy(result, nums)

	// 简单冒泡排序
	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if result[j] > result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return result
}

// SortDesc 降序排序
func SortDesc(nums []int) []int {
	result := make([]int, len(nums))
	copy(result, nums)

	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if result[j] < result[j+1] {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}

	return result
}

// ToString 转字符串
func ToString(n int) string {
	return strconv.Itoa(n)
}

// ToStringFloat 浮点数转字符串
func ToStringFloat(n float64) string {
	return strconv.FormatFloat(n, 'f', -1, 64)
}

// ToInt 字符串转整数
func ToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// ToFloat 字符串转浮点数
func ToFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// ParseHex 解析十六进制
func ParseHex(s string) (int, error) {
	n, err := strconv.ParseInt(s, 16, 64)
	return int(n), err
}

// ParseOctal 解析八进制
func ParseOctal(s string) (int, error) {
	n, err := strconv.ParseInt(s, 8, 64)
	return int(n), err
}

// ParseBinary 解析二进制
func ParseBinary(s string) (int, error) {
	n, err := strconv.ParseInt(s, 2, 64)
	return int(n), err
}

// ToHex 转十六进制
func ToHex(n int) string {
	return strconv.FormatInt(int64(n), 16)
}

// ToOctal 转八进制
func ToOctal(n int) string {
	return strconv.FormatInt(int64(n), 8)
}

// ToBinary 转二进制
func ToBinary(n int) string {
	return strconv.FormatInt(int64(n), 2)
}

// Format 格式化数字
func Format(n int) string {
	s := strconv.Itoa(n)

	var result []byte
	for i := len(s) - 1; i >= 0; i-- {
		pos := len(s) - 1 - i
		if pos > 0 && pos%3 == 0 {
			result = append([]byte{','}, result...)
		}
		result = append([]byte{s[i]}, result...)
	}

	return string(result)
}

// FormatFloat 格式化浮点数
func FormatFloat(n float64, precision int) string {
	format := fmt.Sprintf("%%.%df", precision)
	return fmt.Sprintf(format, n)
}

// FormatCurrency 格式化货币
func FormatCurrency(n float64, symbol string) string {
	formatted := FormatFloat(n, 2)
	return symbol + formatted
}

// IsNumeric 检查字符串是否为数字
func IsNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

// IsInteger 检查字符串是否为整数
func IsInteger(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

// RandomInRange 范围内随机数
func RandomInRange(min, max int) int {
	// 简化实现
	return min + int(float64(max-min)*0.5)
}

// Median 中位数
func Median(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}

	sorted := Sort(nums)
	n := len(sorted)

	if n%2 == 0 {
		return float64(sorted[n/2-1]+sorted[n/2]) / 2
	}

	return float64(sorted[n/2])
}

// Mode 众数
func Mode(nums []int) int {
	if len(nums) == 0 {
		return 0
	}

	counts := make(map[int]int)
	for _, num := range nums {
		counts[num]++
	}

	maxCount := 0
	mode := nums[0]

	for num, count := range counts {
		if count > maxCount {
			maxCount = count
			mode = num
		}
	}

	return mode
}

// Variance 方差
func Variance(nums []int) float64 {
	if len(nums) == 0 {
		return 0
	}

	avg := Average(nums)
	sum := 0.0

	for _, num := range nums {
		diff := float64(num) - avg
		sum += diff * diff
	}

	return sum / float64(len(nums))
}

// StdDev 标准差
func StdDev(nums []int) float64 {
	return math.Sqrt(Variance(nums))
}

// Progress 进度百分比
func Progress(current, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(current) / float64(total) * 100
}

// Lerp 线性插值
func Lerp(start, end, t float64) float64 {
	return start + (end-start)*t
}

// MapRange 映射范围
func MapRange(value, inMin, inMax, outMin, outMax float64) float64 {
	return (value-inMin)*(outMax-outMin)/(inMax-inMin) + outMin
}

// ConstrainAngle 约束角度到0-360
func ConstrainAngle(angle float64) float64 {
	angle = math.Mod(angle, 360)
	if angle < 0 {
		angle += 360
	}
	return angle
}

// Distance 距离
func Distance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// ToRadians 角度转弧度
func ToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// ToDegrees 弧度转角度
func ToDegrees(radians float64) float64 {
	return radians * 180 / math.Pi
}

// NormalizeAngle 标准化角度
func NormalizeAngle(angle float64) float64 {
	for angle < 0 {
		angle += 360
	}
	for angle >= 360 {
		angle -= 360
	}
	return angle
}

// Sign 符号
func Sign(n int) int {
	if n < 0 {
		return -1
	} else if n > 0 {
		return 1
	}
	return 0
}

// SignFloat 浮点数符号
func SignFloat(n float64) float64 {
	if n < 0 {
		return -1
	} else if n > 0 {
		return 1
	}
	return 0
}

// NearlyEqual 近似相等
func NearlyEqual(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

// Interpolate 插值
func Interpolate(a, b, t float64) float64 {
	return a + (b-a)*t
}

// Remap 重新映射
func Remap(value, inputMin, inputMax, outputMin, outputMax float64) float64 {
	return (value-inputMin)/(inputMax-inputMin)*(outputMax-outputMin) + outputMin
}

// BigToInt 大整数转int
func BigToInt(n *big.Int) int {
	return int(n.Int64())
}

// IntToBig int转大整数
func IntToBig(n int) *big.Int {
	return big.NewInt(int64(n))
}

// AddLarge 大数加法
func AddLarge(a, b string) (string, error) {
	intA := new(big.Int)
	intB := new(big.Int)

	_, ok := intA.SetString(a, 10)
	if !ok {
		return "", fmt.Errorf("invalid number: %s", a)
	}

	_, ok = intB.SetString(b, 10)
	if !ok {
		return "", fmt.Errorf("invalid number: %s", b)
	}

	sum := new(big.Int).Add(intA, intB)
	return sum.String(), nil
}

// MulLarge 大数乘法
func MulLarge(a, b string) (string, error) {
	intA := new(big.Int)
	intB := new(big.Int)

	_, ok := intA.SetString(a, 10)
	if !ok {
		return "", fmt.Errorf("invalid number: %s", a)
	}

	_, ok = intB.SetString(b, 10)
	if !ok {
		return "", fmt.Errorf("invalid number: %s", b)
	}

	product := new(big.Int).Mul(intA, intB)
	return product.String(), nil
}
