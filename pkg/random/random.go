// Package random 提供随机数生成工具
package random

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"math/big"
	mrand "math/rand"
	"time"
)

// Int 生成随机整数 [0, max)
func Int(max int) int {
	if max <= 0 {
		return 0
	}
	return mrand.Intn(max)
}

// IntRange 生成随机整数 [min, max)
func IntRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + mrand.Intn(max-min)
}

// Intn 生成随机整数 [0, n)
func Intn(n int) int {
	return mrand.Intn(n)
}

// Int31 生成随机int32
func Int31() int32 {
	return mrand.Int31()
}

// Int31n 生成随机int32 [0, n)
func Int31n(n int32) int32 {
	return mrand.Int31n(n)
}

// Int63 生成随机int64
func Int63() int64 {
	return mrand.Int63()
}

// Int63n 生成随机int64 [0, n)
func Int63n(n int64) int64 {
	return mrand.Int63n(n)
}

// Float64 生成随机浮点数 [0.0, 1.0)
func Float64() float64 {
	return mrand.Float64()
}

// Float64Range 生成随机浮点数 [min, max)
func Float64Range(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + mrand.Float64()*(max-min)
}

// Float32 生成随机float32 [0.0, 1.0)
func Float32() float32 {
	return mrand.Float32()
}

// Bool 生成随机布尔值
func Bool() bool {
	return mrand.Intn(2) == 1
}

// OneOf 从选项中随机选择一个
func OneOf[T any](options []T) T {
	if len(options) == 0 {
		var zero T
		return zero
	}
	return options[mrand.Intn(len(options))]
}

// NOf 从选项中随机选择n个（不重复）
func NOf[T any](options []T, n int) []T {
	if n <= 0 || len(options) == 0 {
		return []T{}
	}
	if n >= len(options) {
		shuffled := make([]T, len(options))
		copy(shuffled, options)
		mrand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		return shuffled
	}

	indices := mrand.Perm(len(options))
	result := make([]T, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, options[indices[i]])
	}
	return result
}

// String 生成随机字符串
func String(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return StringFromCharset(length, charset)
}

// StringLower 生成小写随机字符串
func StringLower(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	return StringFromCharset(length, charset)
}

// StringUpper 生成大写随机字符串
func StringUpper(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return StringFromCharset(length, charset)
}

// StringAlpha 生成字母随机字符串
func StringAlpha(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	return StringFromCharset(length, charset)
}

// StringNumeric 生成数字随机字符串
func StringNumeric(length int) string {
	const charset = "0123456789"
	return StringFromCharset(length, charset)
}

// StringFromCharset 从指定字符集生成随机字符串
func StringFromCharset(length int, charset string) string {
	if length <= 0 {
		return ""
	}

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[mrand.Intn(len(charset))]
	}
	return string(result)
}

// Bytes 生成随机字节
func Bytes(n int) []byte {
	if n <= 0 {
		return []byte{}
	}

	result := make([]byte, n)
	_, err := mrand.Read(result)
	if err != nil {
		// Fallback to crypto/rand
		_, err = cryptoRandRead(result)
		if err != nil {
			panic(fmt.Sprintf("failed to generate random bytes: %v", err))
		}
	}
	return result
}

// BytesSecure 生成安全的随机字节
func BytesSecure(n int) []byte {
	if n <= 0 {
		return []byte{}
	}

	result := make([]byte, n)
	_, err := cryptoRandRead(result)
	if err != nil {
		panic(fmt.Sprintf("failed to generate secure random bytes: %v", err))
	}
	return result
}

// cryptoRandRead 使用crypto/rand读取随机字节
func cryptoRandRead(b []byte) (int, error) {
	n, err := rand.Read(b)
	return n, err
}

// UUID 生成UUID v4
func UUID() string {
	b := BytesSecure(16)
	b[6] = (b[6] & 0x0f) | 0x40 // 版本4
	b[8] = (b[8] & 0x3f) | 0x80 // 变体

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// ULID 生成ULID（简化实现）
func ULID() string {
	t := time.Now().UnixMilli()
	b := BytesSecure(10)

	// 组合时间戳和随机字节
	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data[:8], uint64(t))
	copy(data[8:], b)

	return fmt.Sprintf("%x%04x%04x%04x%04x%04x%04x%04x",
		data[0:4], data[4:5], data[5:6], data[6:7],
		data[7:8], data[8:9], data[9:10], data[10:11])
}

// Hex 生成随机十六进制字符串
func Hex(length int) string {
	if length <= 0 {
		return ""
	}

	// 每个字节转换为2个十六进制字符
	n := (length + 1) / 2
	b := BytesSecure(n)
	return fmt.Sprintf("%x", b)[:length]
}

// Base64 生成随机Base64字符串
func Base64(length int) string {
	if length <= 0 {
		return ""
	}

	// Base64编码会使长度增加约1/3，所以我们需要更少的字节
	n := (length * 3) / 4
	if n == 0 {
		n = 1
	}

	b := BytesSecure(n)
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	result := make([]byte, 0)

	for _, byte := range b {
		result = append(result, charset[byte%64])
	}

	if len(result) > length {
		result = result[:length]
	}

	return string(result)
}

// Shuffle 打乱切片
func Shuffle[T any](slice []T) []T {
	if len(slice) == 0 {
		return slice
	}

	result := make([]T, len(slice))
	copy(result, slice)

	mrand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}

// Sample 随机采样（可重复）
func Sample[T any](slice []T, n int) []T {
	if n <= 0 || len(slice) == 0 {
		return []T{}
	}

	result := make([]T, n)
	for i := 0; i < n; i++ {
		result[i] = slice[mrand.Intn(len(slice))]
	}
	return result
}

// SampleUnique 随机采样（不重复）
func SampleUnique[T any](slice []T, n int) []T {
	if n <= 0 || len(slice) == 0 {
		return []T{}
	}
	if n >= len(slice) {
		return Shuffle(slice)
	}

	indices := mrand.Perm(len(slice))
	result := make([]T, 0, n)
	for i := 0; i < n; i++ {
		result = append(result, slice[indices[i]])
	}
	return result
}

// Perm 生成排列
func Perm(n int) []int {
	return mrand.Perm(n)
}

// Weighted 权重随机选择
func Weighted[T any](items []T, weights []int) (T, error) {
	if len(items) == 0 || len(items) != len(weights) {
		var zero T
		return zero, fmt.Errorf("invalid items or weights")
	}

	total := 0
	for _, w := range weights {
		if w < 0 {
			var zero T
			return zero, fmt.Errorf("negative weight")
		}
		total += w
	}

	if total == 0 {
		var zero T
		return zero, fmt.Errorf("total weight is zero")
	}

	r := Int(total)
	cumsum := 0
	for i, w := range weights {
		cumsum += w
		if r < cumsum {
			return items[i], nil
		}
	}

	return items[len(items)-1], nil
}

// WeightedSample 权重随机采样（不重复）
func WeightedSample[T any](items []T, weights []int, n int) ([]T, error) {
	if n <= 0 || len(items) == 0 {
		return []T{}, nil
	}
	if n >= len(items) {
		return items, nil
	}

	result := make([]T, 0, n)
	remainingIndices := make([]int, len(items))
	remainingWeights := make([]int, len(weights))

	for i := range items {
		remainingIndices[i] = i
		remainingWeights[i] = weights[i]
	}

	for len(result) < n && len(remainingIndices) > 0 {
		total := 0
		for _, w := range remainingWeights {
			total += w
		}

		if total == 0 {
			break
		}

		r := Int(total)
		cumsum := 0
		selectedIdx := 0

		for i, w := range remainingWeights {
			cumsum += w
			if r < cumsum {
				selectedIdx = i
				break
			}
		}

		result = append(result, items[remainingIndices[selectedIdx]])

		// 移除已选项
		remainingIndices = append(remainingIndices[:selectedIdx], remainingIndices[selectedIdx+1:]...)
		remainingWeights = append(remainingWeights[:selectedIdx], remainingWeights[selectedIdx+1:]...)
	}

	return result, nil
}

// Prime 生成质数（简化实现）
func Prime(bits int) (int64, error) {
	if bits <= 0 || bits > 64 {
		return 0, fmt.Errorf("invalid bits: %d", bits)
	}

	for {
		n, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), uint(bits)))
		if err != nil {
			return 0, err
		}

		candidate := n.Int64()
		if isPrime(candidate) {
			return candidate, nil
		}
	}
}

// isPrime 检查是否为质数
func isPrime(n int64) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}

	for i := int64(5); i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// Date 生成随机日期
func Date(min, max time.Time) time.Time {
	if min.After(max) {
		return min
	}

	delta := max.Unix() - min.Unix()
	if delta <= 0 {
		return min
	}

	offset := Int63n(delta + 1)
	return min.Add(time.Duration(offset) * time.Second)
}

// DateRange 生成日期范围内的随机日期
func DateRange(start, end time.Time, n int) []time.Time {
	if n <= 0 {
		return []time.Time{}
	}

	dates := make([]time.Time, n)
	for i := 0; i < n; i++ {
		dates[i] = Date(start, end)
	}
	return dates
}

// IP 生成随机IP地址
func IP() string {
	return fmt.Sprintf("%d.%d.%d.%d",
		ByteRange(1, 255),
		ByteRange(0, 255),
		ByteRange(0, 255),
		ByteRange(0, 255))
}

// IPv6 生成随机IPv6地址（简化实现）
func IPv6() string {
	b := Bytes(16)
	return fmt.Sprintf("%x:%x:%x:%x:%x:%x:%x:%x",
		binary.BigEndian.Uint16(b[0:2]),
		binary.BigEndian.Uint16(b[2:4]),
		binary.BigEndian.Uint16(b[4:6]),
		binary.BigEndian.Uint16(b[6:8]),
		binary.BigEndian.Uint16(b[8:10]),
		binary.BigEndian.Uint16(b[10:12]),
		binary.BigEndian.Uint16(b[12:14]),
		binary.BigEndian.Uint16(b[14:16]))
}

// MAC 生成随机MAC地址
func MAC() string {
	b := Bytes(6)
	// 设置本地 administered 位
	b[0] |= 0x02
	return fmt.Sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		b[0], b[1], b[2], b[3], b[4], b[5])
}

// Byte 生成随机字节
func Byte() byte {
	return byte(mrand.Intn(256))
}

// ByteRange 生成指定范围内的随机字节
func ByteRange(min, max byte) byte {
	if min >= max {
		return min
	}
	return min + byte(mrand.Intn(int(max-min)))
}

// Lorem 生成Lorem ipsum文本
func Lorem(wordCount int) string {
	words := []string{
		"lorem", "ipsum", "dolor", "sit", "amet",
		"consectetur", "adipiscing", "elit", "sed", "do",
		"eiusmod", "tempor", "incididunt", "ut", "labore",
		"et", "dolore", "magna", "aliqua", "ut", "enim",
		"ad", "minim", "veniam", "quis", "nostrud",
		"exercitation", "ullamco", "laboris", "nisi", "ut",
		"aliquip", "ex", "ea", "commodo", "consequat",
	}

	if wordCount <= 0 {
		return ""
	}

	result := make([]string, 0, wordCount)
	for i := 0; i < wordCount; i++ {
		result = append(result, words[mrand.Intn(len(words))])
	}

	text := result[0]
	for i := 1; i < len(result); i++ {
		text += " " + result[i]
	}

	return text
}

// Paragraph 生成随机段落
func Paragraph(sentenceCount int) string {
	if sentenceCount <= 0 {
		return ""
	}

	paragraph := ""
	for i := 0; i < sentenceCount; i++ {
		words := Lorem(IntRange(5, 15))
		paragraph += string(words[0]-'a'+'A') + words[1:] + ". "
	}

	return paragraph[:len(paragraph)-1]
}

// Email 生成随机邮箱
func Email() string {
	domains := []string{"example.com", "test.com", "demo.com", "sample.com"}
	username := StringLower(IntRange(5, 12))
	domain := OneOf(domains)
	return username + "@" + domain
}

// Username 生成随机用户名
func Username() string {
	adjectives := []string{"happy", "silly", "jolly", "quick", "swift", "brave", "calm", "eager"}
	nouns := []string{"fox", "bear", "eagle", "tiger", "lion", "wolf", "hawk", "shark"}

	return OneOf(adjectives) + OneOf(nouns) + StringNumeric(IntRange(1, 4))
}

// Password 生成随机密码
func Password(length int, includeUpper, includeLower, includeDigits, includeSymbols bool) string {
	if length <= 0 {
		return ""
	}

	var charset string
	if includeUpper {
		charset += "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	if includeLower {
		charset += "abcdefghijklmnopqrstuvwxyz"
	}
	if includeDigits {
		charset += "0123456789"
	}
	if includeSymbols {
		charset += "!@#$%^&*()_+-=[]{}|;:,.<>?"
	}

	if charset == "" {
		charset = "abcdefghijklmnopqrstuvwxyz"
	}

	return StringFromCharset(length, charset)
}

// Phone 生成随机手机号（中国大陆）
func Phone() string {
	prefixes := []string{"130", "131", "132", "133", "135", "136", "137", "138", "139",
		"150", "151", "152", "153", "155", "156", "157", "158", "159",
		"180", "181", "182", "183", "185", "186", "187", "188", "189"}

	prefix := OneOf(prefixes)
	suffix := StringNumeric(8)
	return prefix + suffix
}

// IDCard 生成随机身份证号（中国大陆，简化实现）
func IDCard() string {
	// 18位身份证
	areaCode := StringNumeric(6)
	birthday := time.Date(
		IntRange(1950, 2005),
		time.Month(IntRange(1, 13)),
		IntRange(1, 29),
		0, 0, 0, 0, time.UTC,
	).Format("20060102")

	sequenceCode := StringNumeric(3)

	// 简化的校验码计算
	checkCode := string('0' + ByteRange(0, 10))

	return areaCode + birthday + sequenceCode + checkCode
}

// CreditCard 生成随机信用卡号（Luhn算法）
func CreditCard() string {
	prefixes := []string{"4", "51", "52", "53", "54", "55"}
	prefix := OneOf(prefixes)
	length := 16

	// 生成前15位
	number := prefix
	for len(number) < length-1 {
		number += string('0' + ByteRange(0, 10))
	}

	// 计算校验位（Luhn算法）
	checkDigit := calculateLuhnCheckDigit(number)
	return number + string(checkDigit)
}

// calculateLuhnCheckDigit 计算Luhn校验位
func calculateLuhnCheckDigit(number string) byte {
	sum := 0
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		alternate = !alternate
	}

	checkDigit := (10 - (sum % 10)) % 10
	return byte('0' + checkDigit)
}

// Color 生成随机颜色（十六进制）
func Color() string {
	return fmt.Sprintf("#%06x", Int63n(0x1000000))
}

// RGB 生成随机RGB颜色
func RGB() (r, g, b byte) {
	return Byte(), Byte(), Byte()
}

// HSL 生成随机HSL颜色
func HSL() (h, s, l float64) {
	h = Float64() * 360
	s = Float64()
	l = Float64()
	return
}

// LatLong 生成随机经纬度
func LatLong() (lat, long float64) {
	lat = Float64Range(-90, 90)
	long = Float64Range(-180, 180)
	return
}

// Country 生成随机国家代码
func Country() string {
	countries := []string{"US", "CN", "JP", "DE", "GB", "FR", "IN", "IT", "CA", "BR"}
	return OneOf(countries)
}

// Language 生成随机语言代码
func Language() string {
	languages := []string{"en", "zh", "es", "fr", "de", "ja", "ko", "ru", "ar", "pt"}
	return OneOf(languages)
}

// TimeZone 生成随机时区
func TimeZone() string {
	timeZones := []string{
		"America/New_York", "America/Los_Angeles", "America/Chicago",
		"Europe/London", "Europe/Paris", "Europe/Berlin",
		"Asia/Shanghai", "Asia/Tokyo", "Asia/Seoul",
		"Australia/Sydney", "Pacific/Auckland",
	}
	return OneOf(timeZones)
}

// UserAgent 生成随机User-Agent
func UserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36",
		"Mozilla/5.0 (iPhone; CPU iPhone OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15",
		"Mozilla/5.0 (iPad; CPU OS 14_7_1 like Mac OS X) AppleWebKit/605.1.15",
	}
	return OneOf(userAgents)
}

// init 初始化随机种子
func init() {
	mrand.Seed(time.Now().UnixNano())
}
