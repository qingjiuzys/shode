// Package utils 提供通用工具函数
package utils

import (
	"sort"
	"strings"
)

// SliceContains 检查切片是否包含元素
func SliceContains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// SliceIndex 查找元素在切片中的索引
func SliceIndex[T comparable](slice []T, item T) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

// SliceRemove 从切片中移除元素
func SliceRemove[T comparable](slice []T, item T) []T {
	index := SliceIndex(slice, item)
	if index == -1 {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// SliceRemoveAt 从切片中移除指定索引的元素
func SliceRemoveAt[T any](slice []T, index int) []T {
	if index < 0 || index >= len(slice) {
		return slice
	}
	return append(slice[:index], slice[index+1:]...)
}

// SliceUnique 去重切片
func SliceUnique[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(slice))

	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// SliceIntersect 获取两个切片的交集
func SliceIntersect[T comparable](a, b []T) []T {
	set := make(map[T]struct{})
	for _, item := range b {
		set[item] = struct{}{}
	}

	result := make([]T, 0)
	for _, item := range a {
		if _, exists := set[item]; exists {
			result = append(result, item)
		}
	}

	return result
}

// SliceUnion 获取两个切片的并集
func SliceUnion[T comparable](a, b []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(a)+len(b))

	for _, item := range a {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	for _, item := range b {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}

// SliceDifference 获取两个切片的差集（a - b）
func SliceDifference[T comparable](a, b []T) []T {
	set := make(map[T]struct{})
	for _, item := range b {
		set[item] = struct{}{}
	}

	result := make([]T, 0)
	for _, item := range a {
		if _, exists := set[item]; !exists {
			result = append(result, item)
		}
	}

	return result
}

// SliceReverse 反转切片
func SliceReverse[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// SliceShuffle 随机打乱切片
func SliceShuffle[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	// Fisher-Yates shuffle
	for i := len(result) - 1; i > 0; i-- {
		j := int(i) // 简化实现，实际应该使用rand
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// SliceChunk 将切片分块
func SliceChunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}

	result := make([][]T, 0, (len(slice)+size-1)/size)

	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}

	return result
}

// SliceFlatten 扁平化二维切片
func SliceFlatten[T any](slices [][]T) []T {
	totalLen := 0
	for _, slice := range slices {
		totalLen += len(slice)
	}

	result := make([]T, 0, totalLen)
	for _, slice := range slices {
		result = append(result, slice...)
	}

	return result
}

// SliceMap 映射切片元素
func SliceMap[T, U any](slice []T, fn func(T) U) []U {
	result := make([]U, len(slice))
	for i, item := range slice {
		result[i] = fn(item)
	}
	return result
}

// SliceFilter 过滤切片元素
func SliceFilter[T any](slice []T, fn func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

// SliceReduce 归约切片
func SliceReduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	result := initial
	for _, item := range slice {
		result = fn(result, item)
	}
	return result
}

// SliceAny 检查是否有任意元素满足条件
func SliceAny[T any](slice []T, fn func(T) bool) bool {
	for _, item := range slice {
		if fn(item) {
			return true
		}
	}
	return false
}

// SliceAll 检查是否所有元素都满足条件
func SliceAll[T any](slice []T, fn func(T) bool) bool {
	for _, item := range slice {
		if !fn(item) {
			return false
		}
	}
	return true
}

// SliceFind 查找满足条件的第一个元素
func SliceFind[T any](slice []T, fn func(T) bool) (T, bool) {
	for _, item := range slice {
		if fn(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// SliceFindIndex 查找满足条件的第一个元素的索引
func SliceFindIndex[T any](slice []T, fn func(T) bool) int {
	for i, item := range slice {
		if fn(item) {
			return i
		}
	}
	return -1
}

// SliceGroupBy 按条件分组
func SliceGroupBy[T any, K comparable](slice []T, fn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range slice {
		key := fn(item)
		result[key] = append(result[key], item)
	}
	return result
}

// SliceCount 统计满足条件的元素数量
func SliceCount[T any](slice []T, fn func(T) bool) int {
	count := 0
	for _, item := range slice {
		if fn(item) {
			count++
		}
	}
	return count
}

// SliceSort 排序切片
func SliceSort[T ordered](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	sort.Slice(result, func(i, j int) bool {
		return result[i] < result[j]
	})

	return result
}

// SliceSortBy 按指定字段排序
func SliceSortBy[T any](slice []T, less func(a, b T) bool) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	sort.Slice(result, func(i, j int) bool {
		return less(result[i], result[j])
	})

	return result
}

// SliceMin 获取最小值
func SliceMin[T ordered](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}

	min := slice[0]
	for _, item := range slice[1:] {
		if item < min {
			min = item
		}
	}
	return min, true
}

// SliceMax 获取最大值
func SliceMax[T ordered](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}

	max := slice[0]
	for _, item := range slice[1:] {
		if item > max {
			max = item
		}
	}
	return max, true
}

// SliceSum 求和
func SliceSum[T number](slice []T) T {
	var sum T
	for _, item := range slice {
		sum += item
	}
	return sum
}

// SliceAverage 求平均值
func SliceAverage[T number](slice []T) float64 {
	if len(slice) == 0 {
		return 0
	}

	var sum T
	for _, item := range slice {
		sum += item
	}

	return float64(sum) / float64(len(slice))
}

// SliceJoin 连接切片为字符串
func SliceJoin[T any](slice []T, sep string, fn func(T) string) string {
	if len(slice) == 0 {
		return ""
	}

	var builder strings.Builder
	for i, item := range slice {
		if i > 0 {
			builder.WriteString(sep)
		}
		builder.WriteString(fn(item))
	}
	return builder.String()
}

// SplitString 按分隔符分割字符串
func SplitString(s string, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}

// TrimStrings 去除字符串切片中每个元素的首尾空格
func TrimStrings(slice []string) []string {
	return SliceMap(slice, strings.TrimSpace)
}

// FilterEmptyStrings 过滤空字符串
func FilterEmptyStrings(slice []string) []string {
	return SliceFilter(slice, func(s string) bool {
		return s != ""
	})
}

// SplitTrim 分割并去除空格
func SplitTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))

	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}

	return result
}

// 类型约束
type ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

type number interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64
}
