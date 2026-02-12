package semver

import (
	"fmt"
	"strconv"
	"strings"
)

// Version 语义化版本
type Version struct {
	Major int
	Minor int
	Patch int
	Pre   string
	Build string
}

// Parse 解析版本字符串
func Parse(version string) (*Version, error) {
	v := &Version{}

	// 处理通配符版本（如 1.x, 2.*）
	if strings.Contains(version, "x") || strings.Contains(version, "*") {
		// 将通配符转换为 0，用于比较
		version = strings.ReplaceAll(version, "x", "0")
		version = strings.ReplaceAll(version, "*", "0")
	}

	// 分离版本号和构建信息
	parts := strings.SplitN(version, "+", 2)
	if len(parts) == 2 {
		v.Build = parts[1]
		version = parts[0]
	}

	// 分离预发布版本
	parts = strings.SplitN(version, "-", 2)
	if len(parts) == 2 {
		v.Pre = parts[1]
		version = parts[0]
	}

	// 解析主版本号
	numbers := strings.Split(version, ".")
	if len(numbers) < 3 {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	major, err := strconv.Atoi(numbers[0])
	if err != nil {
		return nil, fmt.Errorf("invalid major version: %s", numbers[0])
	}

	minor, err := strconv.Atoi(numbers[1])
	if err != nil {
		return nil, fmt.Errorf("invalid minor version: %s", numbers[1])
	}

	patch, err := strconv.Atoi(numbers[2])
	if err != nil {
		return nil, fmt.Errorf("invalid patch version: %s", numbers[2])
	}

	v.Major = major
	v.Minor = minor
	v.Patch = patch

	return v, nil
}

// String 返回版本字符串
func (v *Version) String() string {
	version := fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.Pre != "" {
		version += "-" + v.Pre
	}
	if v.Build != "" {
		version += "+" + v.Build
	}
	return version
}

// Compare 比较版本
// 返回值: -1 表示 v < other, 0 表示 v == other, 1 表示 v > other
func (v *Version) Compare(other *Version) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}

	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}

	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}

	return 0
}

// GreaterThan 检查是否大于
func (v *Version) GreaterThan(other *Version) bool {
	return v.Compare(other) > 0
}

// LessThan 检查是否小于
func (v *Version) LessThan(other *Version) bool {
	return v.Compare(other) < 0
}

// Equals 检查是否相等
func (v *Version) Equals(other *Version) bool {
	return v.Compare(other) == 0
}

// MustParseVersion 解析版本字符串，失败时 panic
func MustParseVersion(version string) *Version {
	v, err := Parse(version)
	if err != nil {
		panic(err)
	}
	return v
}

// ParseVersion 解析版本字符串（别名）
func ParseVersion(version string) (*Version, error) {
	return Parse(version)
}

// Range 版本范围
type Range struct {
	min *Version
	max *Version
}

// ParseRange 解析版本范围
func ParseRange(rangeStr string) (*Range, error) {
	// 简化实现，支持 "^1.2.3", "~1.2.3", ">=1.2.3", "1.2.3 - 2.0.0"
	rangeStr = strings.TrimSpace(rangeStr)

	if strings.HasPrefix(rangeStr, "^") {
		// 兼容性版本
		version, err := Parse(rangeStr[1:])
		if err != nil {
			return nil, err
		}
		return &Range{
			min: version,
			max: &Version{Major: version.Major + 1, Minor: 0, Patch: 0},
		}, nil
	}

	if strings.HasPrefix(rangeStr, "~") {
		// 补丁版本
		version, err := Parse(rangeStr[1:])
		if err != nil {
			return nil, err
		}
		return &Range{
			min: version,
			max: &Version{Major: version.Major, Minor: version.Minor + 1, Patch: 0},
		}, nil
	}

	if strings.Contains(rangeStr, " - ") {
		// 范围
		parts := strings.Split(rangeStr, " - ")
		min, err := Parse(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, err
		}
		max, err := Parse(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		return &Range{min: min, max: max}, nil
	}

	// 精确版本
	version, err := Parse(rangeStr)
	if err != nil {
		return nil, err
	}
	return &Range{min: version, max: version}, nil
}

// Contains 检查版本是否在范围内
func (r *Range) Contains(v *Version) bool {
	if v == nil {
		return false
	}
	cmpMin := v.Compare(r.min)
	cmpMax := v.Compare(r.max)
	return cmpMin >= 0 && cmpMax <= 0
}

// Match 检查版本是否匹配（别名）
func (r *Range) Match(v *Version) bool {
	return r.Contains(v)
}

// Intersect 计算两个范围的交集
func (r *Range) Intersect(other *Range) *Range {
	if r == nil || other == nil {
		return nil
	}

	min := r.min
	if other.min.GreaterThan(min) {
		min = other.min
	}

	max := r.max
	if other.max.LessThan(max) {
		max = other.max
	}

	if min.GreaterThan(max) {
		return nil
	}

	return &Range{min: min, max: max}
}

// MaxSatisfying 返回范围内满足条件的最大版本
func (r *Range) MaxSatisfying(versions []*Version) *Version {
	var maxVer *Version
	for _, v := range versions {
		if r.Contains(v) {
			if maxVer == nil || v.GreaterThan(maxVer) {
				maxVer = v
			}
		}
	}
	return maxVer
}
