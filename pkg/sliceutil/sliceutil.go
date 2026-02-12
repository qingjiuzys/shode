// Package sliceutil 提供切片操作工具
package sliceutil

import (
	"fmt"
	"math/rand"
	"strings"
)

// Ordered 可排序的类型约束
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

// Contains 检查切片是否包含元素
func Contains[T comparable](slice []T, item T) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

// Index 查找元素索引
func Index[T comparable](slice []T, item T) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

// LastIndex 查找元素最后出现的索引
func LastIndex[T comparable](slice []T, item T) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] == item {
			return i
		}
	}
	return -1
}

// Find 查找满足条件的元素
func Find[T any](slice []T, predicate func(T) bool) (T, bool) {
	var zero T
	for _, v := range slice {
		if predicate(v) {
			return v, true
		}
	}
	return zero, false
}

// Filter 过滤切片
func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range slice {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Map 映射切片
func Map[T any, U any](slice []T, mapper func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = mapper(v)
	}
	return result
}

// Reduce 归约切片
func Reduce[T any, U any](slice []T, initial U, reducer func(U, T) U) U {
	result := initial
	for _, v := range slice {
		result = reducer(result, v)
	}
	return result
}

// ForEach 遍历切片
func ForEach[T any](slice []T, fn func(T)) {
	for _, v := range slice {
		fn(v)
	}
}

// ForEachWithIndex 遍历切片（带索引）
func ForEachWithIndex[T any](slice []T, fn func(int, T)) {
	for i, v := range slice {
		fn(i, v)
	}
}

// Any 是否有元素满足条件
func Any[T any](slice []T, predicate func(T) bool) bool {
	for _, v := range slice {
		if predicate(v) {
			return true
		}
	}
	return false
}

// All 是否所有元素满足条件
func All[T any](slice []T, predicate func(T) bool) bool {
	for _, v := range slice {
		if !predicate(v) {
			return false
		}
	}
	return true
}

// None 是否没有元素满足条件
func None[T any](slice []T, predicate func(T) bool) bool {
	return !Any(slice, predicate)
}

// Chunk 分块
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return [][]T{slice}
	}

	var result [][]T
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}
	return result
}

// Flatten 扁平化
func Flatten[T any](slice [][]T) []T {
	var result []T
	for _, inner := range slice {
		result = append(result, inner...)
	}
	return result
}

// Reverse 反转切片
func Reverse[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result
}

// Unique 去重
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0)

	for _, v := range slice {
		if _, exists := seen[v]; !exists {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}

	return result
}

// Union 并集
func Union[T comparable](slices ...[]T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0)

	for _, slice := range slices {
		for _, v := range slice {
			if _, exists := seen[v]; !exists {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}

	return result
}

// Intersection 交集
func Intersection[T comparable](slices ...[]T) []T {
	if len(slices) == 0 {
		return []T{}
	}

	// 统计每个元素出现的次数
	counts := make(map[T]int)
	for _, slice := range slices {
		unique := Unique(slice)
		for _, v := range unique {
			counts[v]++
		}
	}

	// 只保留在所有切片中都出现的元素
	result := make([]T, 0)
	for v, count := range counts {
		if count == len(slices) {
			result = append(result, v)
		}
	}

	return result
}

// Difference 差集
func Difference[T comparable](slice []T, others ...[]T) []T {
	// 创建others的并集
	othersSet := make(map[T]struct{})
	for _, other := range others {
		for _, v := range other {
			othersSet[v] = struct{}{}
		}
	}

	// 过滤掉在othersSet中的元素
	result := make([]T, 0)
	for _, v := range slice {
		if _, exists := othersSet[v]; !exists {
			result = append(result, v)
		}
	}

	return result
}

// Shuffle 打乱切片
func Shuffle[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	rand.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return result
}

// Sample 随机采样
func Sample[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	if n >= len(slice) {
		return slice
	}

	shuffled := Shuffle(slice)
	return shuffled[:n]
}

// Sort 排序
func Sort[T Ordered](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	// 简单冒泡排序
	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if compareOrdered(result[j], result[j+1]) > 0 {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

// SortBy 按条件排序
func SortBy[T any](slice []T, less func(a, b T) bool) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if less(result[j+1], result[j]) {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

// SortDesc 降序排序
func SortDesc[T Ordered](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	for i := 0; i < len(result)-1; i++ {
		for j := 0; j < len(result)-i-1; j++ {
			if compareOrdered(result[j], result[j+1]) < 0 {
				result[j], result[j+1] = result[j+1], result[j]
			}
		}
	}
	return result
}

// Min 获取最小值
func Min[T Ordered](slice []T) (T, error) {
	var zero T
	if len(slice) == 0 {
		return zero, fmt.Errorf("slice is empty")
	}
	min := slice[0]
	for _, v := range slice[1:] {
		if compareOrdered(v, min) < 0 {
			min = v
		}
	}
	return min, nil
}

// Max 获取最大值
func Max[T Ordered](slice []T) (T, error) {
	var zero T
	if len(slice) == 0 {
		return zero, fmt.Errorf("slice is empty")
	}
	max := slice[0]
	for _, v := range slice[1:] {
		if compareOrdered(v, max) > 0 {
			max = v
		}
	}
	return max, nil
}

// compareOrdered 比较两个有序值
func compareOrdered[T Ordered](a, b T) int {
	switch any(a).(type) {
	case int:
		ai, bi := any(a).(int), any(b).(int)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case int8:
		ai, bi := any(a).(int8), any(b).(int8)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case int16:
		ai, bi := any(a).(int16), any(b).(int16)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case int32:
		ai, bi := any(a).(int32), any(b).(int32)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case int64:
		ai, bi := any(a).(int64), any(b).(int64)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case uint:
		ai, bi := any(a).(uint), any(b).(uint)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case uint8:
		ai, bi := any(a).(uint8), any(b).(uint8)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case uint16:
		ai, bi := any(a).(uint16), any(b).(uint16)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case uint32:
		ai, bi := any(a).(uint32), any(b).(uint32)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case uint64:
		ai, bi := any(a).(uint64), any(b).(uint64)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case uintptr:
		ai, bi := any(a).(uintptr), any(b).(uintptr)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case float32:
		ai, bi := any(a).(float32), any(b).(float32)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case float64:
		ai, bi := any(a).(float64), any(b).(float64)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	case string:
		ai, bi := any(a).(string), any(b).(string)
		if ai < bi {
			return -1
		} else if ai > bi {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// Sum 求和
func Sum[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](slice []T) T {
	var sum T
	for _, v := range slice {
		sum += v
	}
	return sum
}

// Average 求平均值
func Average[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](slice []T) float64 {
	if len(slice) == 0 {
		return 0
	}
	sum := Sum(slice)
	return float64(sum) / float64(len(slice))
}

// Product 求积
func Product[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](slice []T) T {
	product := T(1)
	for _, v := range slice {
		product *= v
	}
	return product
}

// Count 统计满足条件的元素数量
func Count[T any](slice []T, predicate func(T) bool) int {
	count := 0
	for _, v := range slice {
		if predicate(v) {
			count++
		}
	}
	return count
}

// CountBy 统计元素出现次数
func CountBy[T comparable](slice []T) map[T]int {
	counts := make(map[T]int)
	for _, v := range slice {
		counts[v]++
	}
	return counts
}

// GroupBy 分组
func GroupBy[T any, K comparable](slice []T, keyFunc func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, v := range slice {
		key := keyFunc(v)
		result[key] = append(result[key], v)
	}
	return result
}

// Partition 分区
func Partition[T any](slice []T, predicate func(T) bool) ([]T, []T) {
	matched := make([]T, 0)
	unmatched := make([]T, 0)

	for _, v := range slice {
		if predicate(v) {
			matched = append(matched, v)
		} else {
			unmatched = append(unmatched, v)
		}
	}

	return matched, unmatched
}

// Zip 拉链
func Zip[T any](slices ...[]T) [][]T {
	if len(slices) == 0 {
		return [][]T{}
	}

	minLen := len(slices[0])
	for _, s := range slices {
		if len(s) < minLen {
			minLen = len(s)
		}
	}

	result := make([][]T, minLen)
	for i := 0; i < minLen; i++ {
		row := make([]T, len(slices))
		for j, s := range slices {
			row[j] = s[i]
		}
		result[i] = row
	}

	return result
}

// Unzip 拉链解压
func Unzip[T any](slice [][]T) [][]T {
	if len(slice) == 0 {
		return [][]T{}
	}

	result := make([][]T, len(slice[0]))
	for i := range result {
		result[i] = make([]T, len(slice))
	}

	for i, row := range slice {
		for j, v := range row {
			result[j][i] = v
		}
	}

	return result
}

// Concat 连接切片
func Concat[T any](slices ...[]T) []T {
	var result []T
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}

// Push 追加元素
func Push[T any](slice []T, items ...T) []T {
	return append(slice, items...)
}

// Pop 弹出最后一个元素
func Pop[T any](slice []T) ([]T, T, error) {
	var zero T
	if len(slice) == 0 {
		return slice, zero, fmt.Errorf("slice is empty")
	}
	last := len(slice) - 1
	return slice[:last], slice[last], nil
}

// Unshift 在开头插入元素
func Unshift[T any](slice []T, items ...T) []T {
	return append(items, slice...)
}

// Shift 移除第一个元素
func Shift[T any](slice []T) ([]T, T, error) {
	var zero T
	if len(slice) == 0 {
		return slice, zero, fmt.Errorf("slice is empty")
	}
	return slice[1:], slice[0], nil
}

// Delete 删除指定索引的元素
func Delete[T any](slice []T, index int) ([]T, error) {
	if index < 0 || index >= len(slice) {
		return slice, fmt.Errorf("index out of bounds: %d", index)
	}
	return append(slice[:index], slice[index+1:]...), nil
}

// Insert 插入元素
func Insert[T any](slice []T, index int, items ...T) ([]T, error) {
	if index < 0 || index > len(slice) {
		return slice, fmt.Errorf("index out of bounds: %d", index)
	}
	if len(items) == 0 {
		return slice, nil
	}

	result := make([]T, 0, len(slice)+len(items))
	result = append(result, slice[:index]...)
	result = append(result, items...)
	result = append(result, slice[index:]...)

	return result, nil
}

// Replace 替换元素
func Replace[T comparable](slice []T, old, new T) []T {
	result := make([]T, len(slice))
	copy(result, slice)

	for i, v := range result {
		if v == old {
			result[i] = new
		}
	}

	return result
}

// Repeat 重复元素
func Repeat[T any](item T, n int) []T {
	result := make([]T, n)
	for i := range result {
		result[i] = item
	}
	return result
}

// Rotate 旋转切片
func Rotate[T any](slice []T, n int) []T {
	if len(slice) == 0 {
		return slice
	}

	n = n % len(slice)
	if n < 0 {
		n += len(slice)
	}

	return append(slice[n:], slice[:n]...)
}

// Take 获取前n个元素
func Take[T any](slice []T, n int) []T {
	if n <= 0 {
		return []T{}
	}
	if n >= len(slice) {
		return slice
	}
	return slice[:n]
}

// TakeWhile 获取满足条件的元素直到不满足
func TakeWhile[T any](slice []T, predicate func(T) bool) []T {
	for i, v := range slice {
		if !predicate(v) {
			return slice[:i]
		}
	}
	return slice
}

// Drop 跳过前n个元素
func Drop[T any](slice []T, n int) []T {
	if n <= 0 {
		return slice
	}
	if n >= len(slice) {
		return []T{}
	}
	return slice[n:]
}

// DropWhile 跳过满足条件的元素直到不满足
func DropWhile[T any](slice []T, predicate func(T) bool) []T {
	for i, v := range slice {
		if !predicate(v) {
			return slice[i:]
		}
	}
	return []T{}
}

// Join 连接切片为字符串
func Join[T any](slice []T, sep string) string {
	strs := make([]string, len(slice))
	for i, v := range slice {
		strs[i] = fmt.Sprintf("%v", v)
	}
	return strings.Join(strs, sep)
}

// SplitAt 在指定位置分割切片
func SplitAt[T any](slice []T, index int) ([]T, []T) {
	if index <= 0 {
		return []T{}, slice
	}
	if index >= len(slice) {
		return slice, []T{}
	}
	return slice[:index], slice[index:]
}

// SplitBy 按条件分割切片
func SplitBy[T any](slice []T, predicate func(T) bool) [][]T {
	var result [][]T
	current := make([]T, 0)

	for _, v := range slice {
		if predicate(v) {
			if len(current) > 0 {
				result = append(result, current)
				current = make([]T, 0)
			}
		} else {
			current = append(current, v)
		}
	}

	if len(current) > 0 {
		result = append(result, current)
	}

	return result
}

// Fill 填充切片
func Fill[T any](slice []T, value T) []T {
	for i := range slice {
		slice[i] = value
	}
	return slice
}

// FillRange 填充指定范围
func FillRange[T any](slice []T, value T, start, end int) []T {
	if start < 0 {
		start = 0
	}
	if end > len(slice) {
		end = len(slice)
	}
	for i := start; i < end; i++ {
		slice[i] = value
	}
	return slice
}

// Clone 克隆切片
func Clone[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	return result
}

// Equal 比较切片是否相等
func Equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// EqualFunc 使用函数比较切片是否相等
func EqualFunc[T any](a, b []T, eq func(T, T) bool) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !eq(a[i], b[i]) {
			return false
		}
	}
	return true
}

// Compare 比较切片
func Compare[T Ordered](a, b []T) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if cmp := compareOrdered(a[i], b[i]); cmp != 0 {
			return cmp
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

// CompareFunc 使用函数比较切片
func CompareFunc[T any](a, b []T, cmp func(T, T) int) int {
	minLen := len(a)
	if len(b) < minLen {
		minLen = len(b)
	}
	for i := 0; i < minLen; i++ {
		if c := cmp(a[i], b[i]); c != 0 {
			return c
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return 1
	}
	return 0
}

// IndexFunc 查找满足条件的元素索引
func IndexFunc[T any](slice []T, predicate func(T) bool) int {
	for i, v := range slice {
		if predicate(v) {
			return i
		}
	}
	return -1
}

// LastIndexFunc 查找满足条件的元素最后索引
func LastIndexFunc[T any](slice []T, predicate func(T) bool) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if predicate(slice[i]) {
			return i
		}
	}
	return -1
}

// BinarySearch 二分查找
func BinarySearch[T Ordered](slice []T, target T) int {
	low, high := 0, len(slice)-1
	for low <= high {
		mid := (low + high) / 2
		cmp := compareOrdered(slice[mid], target)
		if cmp == 0 {
			return mid
		} else if cmp < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -1
}

// BinarySearchFunc 使用函数二分查找
func BinarySearchFunc[T any](slice []T, target T, cmp func(T, T) int) int {
	low, high := 0, len(slice)-1
	for low <= high {
		mid := (low + high) / 2
		c := cmp(slice[mid], target)
		if c == 0 {
			return mid
		} else if c < 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return -1
}

// InsertSorted 插入到已排序切片
func InsertSorted[T Ordered](slice []T, item T) []T {
	i := BinarySearch(slice, item)
	if i < 0 {
		i = len(slice)
	}
	result := make([]T, 0, len(slice)+1)
	result = append(result, slice[:i]...)
	result = append(result, item)
	result = append(result, slice[i:]...)
	return result
}

// Remove 删除元素
func Remove[T comparable](slice []T, item T) []T {
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if v != item {
			result = append(result, v)
		}
	}
	return result
}

// RemoveAll 删除所有匹配的元素
func RemoveAll[T comparable](slice []T, items []T) []T {
	removeSet := make(map[T]struct{})
	for _, item := range items {
		removeSet[item] = struct{}{}
	}

	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, exists := removeSet[v]; !exists {
			result = append(result, v)
		}
	}

	return result
}

// Permutations 生成排列
func Permutations[T any](slice []T) [][]T {
	if len(slice) == 0 {
		return [][]T{{}}
	}

	var result [][]T
	for i, v := range slice {
		rest := make([]T, 0, len(slice)-1)
		rest = append(rest, slice[:i]...)
		rest = append(rest, slice[i+1:]...)

		for _, perm := range Permutations(rest) {
			result = append(result, append([]T{v}, perm...))
		}
	}

	return result
}

// Combinations 生成组合
func Combinations[T any](slice []T, r int) [][]T {
	if r > len(slice) || r <= 0 {
		return [][]T{}
	}
	if r == len(slice) {
		return [][]T{slice}
	}
	if r == 1 {
		result := make([][]T, len(slice))
		for i, v := range slice {
			result[i] = []T{v}
		}
		return result
	}

	var result [][]T
	for i := 0; i <= len(slice)-r; i++ {
		for _, combo := range Combinations(slice[i+1:], r-1) {
			result = append(result, append([]T{slice[i]}, combo...))
		}
	}

	return result
}

// CartesianProduct 笛卡尔积
func CartesianProduct[T any](slices ...[]T) [][]T {
	if len(slices) == 0 {
		return [][]T{{}}
	}
	if len(slices) == 1 {
		result := make([][]T, len(slices[0]))
		for i, v := range slices[0] {
			result[i] = []T{v}
		}
		return result
	}

	var result [][]T
	for _, v := range slices[0] {
		for _, product := range CartesianProduct(slices[1:]...) {
			result = append(result, append([]T{v}, product...))
		}
	}

	return result
}
