// Package maputil 提供map操作工具
package maputil

import (
	"fmt"
	"strings"
)

// Ordered 可排序的类型约束
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

// Keys 获取所有键
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values 获取所有值
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Entries 获取所有键值对
func Entries[K comparable, V any](m map[K]V) []Entry[K, V] {
	entries := make([]Entry[K, V], 0, len(m))
	for k, v := range m {
		entries = append(entries, Entry[K, V]{Key: k, Value: v})
	}
	return entries
}

// Entry 键值对
type Entry[K comparable, V any] struct {
	Key   K
	Value V
}

// HasKey 检查是否包含键
func HasKey[K comparable, V any](m map[K]V, key K) bool {
	_, exists := m[key]
	return exists
}

// HasValue 检查是否包含值
func HasValue[K comparable, V comparable](m map[K]V, value V) bool {
	for _, v := range m {
		if v == value {
			return true
		}
	}
	return false
}

// Get 获取值，如果不存在返回默认值
func Get[K comparable, V any](m map[K]V, key K, defaultValue V) V {
	if v, exists := m[key]; exists {
		return v
	}
	return defaultValue
}

// GetOrGet 获取值或通过函数获取
func GetOrGet[K comparable, V any](m map[K]V, key K, fn func() V) V {
	if v, exists := m[key]; exists {
		return v
	}
	return fn()
}

// Set 设置键值对
func Set[K comparable, V any](m map[K]V, key K, value V) {
	m[key] = value
}

// Delete 删除键
func Delete[K comparable, V any](m map[K]V, key K) {
	delete(m, key)
}

// DeleteMultiple 删除多个键
func DeleteMultiple[K comparable, V any](m map[K]V, keys ...K) {
	for _, key := range keys {
		delete(m, key)
	}
}

// Clear 清空map
func Clear[K comparable, V any](m map[K]V) {
	for k := range m {
		delete(m, k)
	}
}

// IsEmpty 检查是否为空
func IsEmpty[K comparable, V any](m map[K]V) bool {
	return len(m) == 0
}

// Size 获取大小
func Size[K comparable, V any](m map[K]V) int {
	return len(m)
}

// Clone 克隆map
func Clone[K comparable, V any](m map[K]V) map[K]V {
	cloned := make(map[K]V, len(m))
	for k, v := range m {
		cloned[k] = v
	}
	return cloned
}

// Merge 合并map
func Merge[K comparable, V any](maps ...map[K]V) map[K]V {
	result := make(map[K]V)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// MergeDeep 深度合并
func MergeDeep[K comparable, V any](dest, src map[K]V, mergeFunc func(V, V) V) {
	for key, srcValue := range src {
		if destValue, exists := dest[key]; exists {
			dest[key] = mergeFunc(destValue, srcValue)
		} else {
			dest[key] = srcValue
		}
	}
}

// Intersect 交集（保留第一个map的值）
func Intersect[K comparable, V any](maps ...map[K]V) map[K]V {
	if len(maps) == 0 {
		return map[K]V{}
	}
	if len(maps) == 1 {
		return Clone(maps[0])
	}

	result := make(map[K]V)
	first := maps[0]

	// 统计每个key出现的次数
	counts := make(map[K]int)
	for _, m := range maps {
		for k := range m {
			counts[k]++
		}
	}

	// 只保留在所有map中都出现的key
	for k := range first {
		if counts[k] == len(maps) {
			result[k] = first[k]
		}
	}

	return result
}

// Diff 差集（在第一个map中但不在其他map中）
func Diff[K comparable, V any](maps ...map[K]V) map[K]V {
	if len(maps) == 0 {
		return map[K]V{}
	}
	if len(maps) == 1 {
		return Clone(maps[0])
	}

	result := Clone(maps[0])

	// 移除在其他map中出现的key
	for _, m := range maps[1:] {
		for k := range m {
			delete(result, k)
		}
	}

	return result
}

// Filter 过滤
func Filter[K comparable, V any](m map[K]V, predicate func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// Map 映射值
func Map[K comparable, V any, R any](m map[K]V, mapper func(K, V) R) map[K]R {
	result := make(map[K]R, len(m))
	for k, v := range m {
		result[k] = mapper(k, v)
	}
	return result
}

// MapKeys 映射键
func MapKeys[K comparable, K2 comparable, V any](m map[K]V, keyMapper func(K) K2) map[K2]V {
	result := make(map[K2]V, len(m))
	for k, v := range m {
		result[keyMapper(k)] = v
	}
	return result
}

// ForEach 遍历
func ForEach[K comparable, V any](m map[K]V, fn func(K, V)) {
	for k, v := range m {
		fn(k, v)
	}
}

// Reduce 归约
func Reduce[K comparable, V any, R any](m map[K]V, initial R, reducer func(R, K, V) R) R {
	result := initial
	for k, v := range m {
		result = reducer(result, k, v)
	}
	return result
}

// Any 是否有元素满足条件
func Any[K comparable, V any](m map[K]V, predicate func(K, V) bool) bool {
	for k, v := range m {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// All 是否所有元素满足条件
func All[K comparable, V any](m map[K]V, predicate func(K, V) bool) bool {
	for k, v := range m {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// None 是否没有元素满足条件
func None[K comparable, V any](m map[K]V, predicate func(K, V) bool) bool {
	return !Any(m, predicate)
}

// Find 查找满足条件的值
func Find[K comparable, V any](m map[K]V, predicate func(K, V) bool) (V, bool) {
	for k, v := range m {
		if predicate(k, v) {
			return v, true
		}
	}
	var zero V
	return zero, false
}

// FindKey 查找满足条件的键
func FindKey[K comparable, V any](m map[K]V, predicate func(K, V) bool) (K, bool) {
	for k, v := range m {
		if predicate(k, v) {
			return k, true
		}
	}
	var zero K
	return zero, false
}

// Invert 反转map的键值
func Invert[K comparable, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K, len(m))
	for k, v := range m {
		result[v] = k
	}
	return result
}

// Flip 反转map（值变为切片）
func Flip[K comparable, V comparable](m map[K]V) map[V][]K {
	result := make(map[V][]K)
	for k, v := range m {
		result[v] = append(result[v], k)
	}
	return result
}

// GroupBy 分组
func GroupBy[K comparable, V any, K2 comparable](slice []V, keyFunc func(V) K2) map[K2][]V {
	result := make(map[K2][]V)
	for _, v := range slice {
		key := keyFunc(v)
		result[key] = append(result[key], v)
	}
	return result
}

// Partition 分区
func Partition[K comparable, V any](m map[K]V, predicate func(K, V) bool) (map[K]V, map[K]V) {
	matched := make(map[K]V)
	unmatched := make(map[K]V)

	for k, v := range m {
		if predicate(k, v) {
			matched[k] = v
		} else {
			unmatched[k] = v
		}
	}

	return matched, unmatched
}

// Separate 分离键和值
func Separate[K comparable, V any](m map[K]V) ([]K, []V) {
	keys := make([]K, 0, len(m))
	values := make([]V, 0, len(m))

	for k, v := range m {
		keys = append(keys, k)
		values = append(values, v)
	}

	return keys, values
}

// FromSlice 从切片创建map
func FromSlice[K comparable, V any](slice []V, keyFunc func(V) K) map[K]V {
	result := make(map[K]V, len(slice))
	for _, v := range slice {
		result[keyFunc(v)] = v
	}
	return result
}

// FromSlices 从两个切片创建map
func FromSlices[K comparable, V any](keys []K, values []V) (map[K]V, error) {
	if len(keys) != len(values) {
		return nil, fmt.Errorf("keys and values have different lengths")
	}

	result := make(map[K]V, len(keys))
	for i, key := range keys {
		result[key] = values[i]
	}

	return result, nil
}

// ToSlice 转为切片
func ToSlice[K comparable, V any](m map[K]V) []V {
	return Values(m)
}

// Update 更新map
func Update[K comparable, V any](dest, src map[K]V) {
	for k, v := range src {
		dest[k] = v
	}
}

// Defaults 设置默认值
func Defaults[K comparable, V any](m map[K]V, defaults map[K]V) {
	for k, v := range defaults {
		if _, exists := m[k]; !exists {
			m[k] = v
		}
	}
}

// Pick 选取指定的键
func Pick[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	result := make(map[K]V)
	for _, key := range keys {
		if v, exists := m[key]; exists {
			result[key] = v
		}
	}
	return result
}

// Omit 排除指定的键
func Omit[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	result := make(map[K]V)
	keySet := make(map[K]struct{})
	for _, key := range keys {
		keySet[key] = struct{}{}
	}

	for k, v := range m {
		if _, exists := keySet[k]; !exists {
			result[k] = v
		}
	}
	return result
}

// Rename 重命名键
func Rename[K comparable, V any](m map[K]V, oldKey, newKey K) error {
	if v, exists := m[oldKey]; exists {
		delete(m, oldKey)
		m[newKey] = v
		return nil
	}
	return fmt.Errorf("key not found: %v", oldKey)
}

// Copy 复制指定的键到新map
func Copy[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	return Pick(m, keys...)
}

// Equal 比较两个map是否相等
func Equal[K comparable, V comparable](a, b map[K]V) bool {
	if len(a) != len(b) {
		return false
	}

	for k, va := range a {
		if vb, exists := b[k]; !exists || va != vb {
			return false
		}
	}

	return true
}

// EqualFunc 使用函数比较两个map
func EqualFunc[K comparable, V any](a, b map[K]V, eq func(V, V) bool) bool {
	if len(a) != len(b) {
		return false
	}

	for k, va := range a {
		vb, exists := b[k]
		if !exists || !eq(va, vb) {
			return false
		}
	}

	return true
}

// Count 统计满足条件的元素数量
func Count[K comparable, V any](m map[K]V, predicate func(K, V) bool) int {
	count := 0
	for k, v := range m {
		if predicate(k, v) {
			count++
		}
	}
	return count
}

// CountBy 统计每个值出现的次数
func CountBy[K comparable, V comparable](m map[K]V) map[V]int {
	counts := make(map[V]int)
	for _, v := range m {
		counts[v]++
	}
	return counts
}

// MinValue 获取最小值
func MinValue[K comparable, V Ordered](m map[K]V) (V, error) {
	var zero V
	if len(m) == 0 {
		return zero, fmt.Errorf("map is empty")
	}

	first := true
	var min V

	for _, v := range m {
		if first || compareOrdered(v, min) < 0 {
			min = v
			first = false
		}
	}

	return min, nil
}

// MaxValue 获取最大值
func MaxValue[K comparable, V Ordered](m map[K]V) (V, error) {
	var zero V
	if len(m) == 0 {
		return zero, fmt.Errorf("map is empty")
	}

	first := true
	var max V

	for _, v := range m {
		if first || compareOrdered(v, max) > 0 {
			max = v
			first = false
		}
	}

	return max, nil
}

// compareOrdered 比较两个有序值
func compareOrdered[V Ordered](a, b V) int {
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

// MinBy 使用函数获取最小值
func MinBy[K comparable, V any](m map[K]V, less func(V, V) bool) (K, V, error) {
	var zeroK K
	var zeroV V

	if len(m) == 0 {
		return zeroK, zeroV, fmt.Errorf("map is empty")
	}

	minKey := Keys(m)[0]
	minValue := m[minKey]

	for k, v := range m {
		if less(v, minValue) {
			minKey = k
			minValue = v
		}
	}

	return minKey, minValue, nil
}

// MaxBy 使用函数获取最大值
func MaxBy[K comparable, V any](m map[K]V, greater func(V, V) bool) (K, V, error) {
	var zeroK K
	var zeroV V

	if len(m) == 0 {
		return zeroK, zeroV, fmt.Errorf("map is empty")
	}

	maxKey := Keys(m)[0]
	maxValue := m[maxKey]

	for k, v := range m {
		if greater(v, maxValue) {
			maxKey = k
			maxValue = v
		}
	}

	return maxKey, maxValue, nil
}

// Sum 对值求和
func Sum[K comparable, V ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](m map[K]V) V {
	var sum V
	for _, v := range m {
		sum += v
	}
	return sum
}

// Average 对值求平均
func Average[K comparable, V ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](m map[K]V) float64 {
	if len(m) == 0 {
		return 0
	}
	sum := Sum(m)
	return float64(sum) / float64(len(m))
}

// Join 连接map为字符串
func Join[K comparable, V any](m map[K]V, entrySep, kvSep string, stringFunc func(V) string) string {
	if len(m) == 0 {
		return ""
	}

	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, fmt.Sprintf("%v%s%s", k, kvSep, stringFunc(v)))
	}

	return strings.Join(parts, entrySep)
}

// Split 分割字符串为map
func Split[K comparable](s string, entrySep, kvSep string, keyFunc func(string) K, valueFunc func(string) string) map[K]string {
	if s == "" {
		return map[K]string{}
	}

	result := map[K]string{}
	entries := strings.Split(s, entrySep)

	for _, entry := range entries {
		parts := strings.SplitN(entry, kvSep, 2)
		if len(parts) == 2 {
			key := keyFunc(strings.TrimSpace(parts[0]))
			value := strings.TrimSpace(parts[1])
			result[key] = valueFunc(value)
		}
	}

	return result
}

// Transform 转换map
func Transform[K comparable, V any, K2 comparable, V2 any](m map[K]V, keyFunc func(K, V) K2, valueFunc func(K, V) V2) map[K2]V2 {
	result := make(map[K2]V2, len(m))
	for k, v := range m {
		result[keyFunc(k, v)] = valueFunc(k, v)
	}
	return result
}

// Chunk 分块
func Chunk[K comparable, V any](m map[K]V, size int) []map[K]V {
	if size <= 0 {
		return []map[K]V{Clone(m)}
	}

	keys := Keys(m)
	result := make([]map[K]V, 0)

	for i := 0; i < len(keys); i += size {
		chunk := make(map[K]V)
		end := i + size
		if end > len(keys) {
			end = len(keys)
		}

		for _, key := range keys[i:end] {
			chunk[key] = m[key]
		}

		result = append(result, chunk)
	}

	return result
}

// Flatten 扁平化嵌套map
func Flatten[K comparable, V any](m map[K]any, sep string) map[string]V {
	result := make(map[string]V)
	flattenHelper("", sep, m, result)
	return result
}

func flattenHelper[K comparable, V any](prefix string, sep string, m map[K]any, result map[string]V) {
	for k, v := range m {
		keyStr := fmt.Sprintf("%v", k)
		var key string
		if prefix != "" {
			key = prefix + sep + keyStr
		} else {
			key = keyStr
		}

		switch val := v.(type) {
		case map[K]any:
			flattenHelper(key, sep, val, result)
		case map[string]any:
			flattenHelper(key, sep, val, result)
		default:
			result[key] = v.(V)
		}
	}
}
