// Package stats 提供统计计算工具
package stats

import (
	"math"
	"sort"
)

// Mean 计算平均值
func Mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// Median 计算中位数
func Median(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	n := len(sorted)
	if n%2 == 0 {
		return (sorted[n/2-1] + sorted[n/2]) / 2
	}
	return sorted[n/2]
}

// Mode 计算众数
func Mode(data []float64) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	freq := make(map[float64]int)
	for _, v := range data {
		freq[v]++
	}

	maxFreq := 0
	for _, f := range freq {
		if f > maxFreq {
			maxFreq = f
		}
	}

	modes := make([]float64, 0)
	for v, f := range freq {
		if f == maxFreq {
			modes = append(modes, v)
		}
	}

	return modes
}

// Variance 计算方差
func Variance(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	mean := Mean(data)
	sum := 0.0
	for _, v := range data {
		diff := v - mean
		sum += diff * diff
	}
	return sum / float64(len(data))
}

// StdDev 计算标准差
func StdDev(data []float64) float64 {
	return math.Sqrt(Variance(data))
}

// Min 计算最小值
func Min(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	min := data[0]
	for _, v := range data[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

// Max 计算最大值
func Max(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	max := data[0]
	for _, v := range data[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

// Range 计算范围
func Range(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	return Max(data) - Min(data)
}

// Sum 计算总和
func Sum(data []float64) float64 {
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum
}

// Percentile 计算百分位数
func Percentile(data []float64, p float64) float64 {
	if len(data) == 0 || p < 0 || p > 1 {
		return 0
	}

	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	index := p * float64(len(sorted)-1)
	if index == float64(int(index)) {
		return sorted[int(index)]
	}

	// 线性插值
	lower := sorted[int(index)]
	upper := sorted[int(index)+1]
	return lower + (upper-lower)*(index-float64(int(index)))
}

// Quartiles 计算四分位数
func Quartiles(data []float64) (q1, q2, q3 float64) {
	q1 = Percentile(data, 0.25)
	q2 = Percentile(data, 0.50)
	q3 = Percentile(data, 0.75)
	return q1, q2, q3
}

// IQR 四分位距
func IQR(data []float64) float64 {
	_, _, q3 := Quartiles(data)
	q1, _, _ := Quartiles(data)
	return q3 - q1
}

// Covariance 协方差
func Covariance(data1, data2 []float64) float64 {
	if len(data1) != len(data2) || len(data1) == 0 {
		return 0
	}

	mean1 := Mean(data1)
	mean2 := Mean(data2)

	sum := 0.0
	for i := range data1 {
		sum += (data1[i] - mean1) * (data2[i] - mean2)
	}

	return sum / float64(len(data1))
}

// Correlation 相关系数
func Correlation(data1, data2 []float64) float64 {
	if len(data1) != len(data2) || len(data1) == 0 {
		return 0
	}

	cov := Covariance(data1, data2)
	std1 := StdDev(data1)
	std2 := StdDev(data2)

	if std1 == 0 || std2 == 0 {
		return 0
	}

	return cov / (std1 * std2)
}

// MovingAverage 移动平均
func MovingAverage(data []float64, window int) []float64 {
	if len(data) == 0 || window <= 0 || window > len(data) {
		return []float64{}
	}

	result := make([]float64, len(data)-window+1)
	for i := 0; i < len(result); i++ {
		sum := 0.0
		for j := 0; j < window; j++ {
			sum += data[i+j]
		}
		result[i] = sum / float64(window)
	}

	return result
}

// CumulativeSum 累积和
func CumulativeSum(data []float64) []float64 {
	if len(data) == 0 {
		return []float64{}
	}

	result := make([]float64, len(data))
	sum := 0.0

	for i, v := range data {
		sum += v
		result[i] = sum
	}

	return result
}

// Summary 统计摘要
type Summary struct {
	Count    int     `json:"count"`
	Mean     float64 `json:"mean"`
	Median   float64 `json:"median"`
	StdDev   float64 `json:"std_dev"`
	Variance float64 `json:"variance"`
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`
	Range    float64 `json:"range"`
	Q1       float64 `json:"q1"`
	Q2       float64 `json:"q2"`
	Q3       float64 `json:"q3"`
	IQR      float64 `json:"iqr"`
}

// Summarize 生成统计摘要
func Summarize(data []float64) *Summary {
	if len(data) == 0 {
		return &Summary{}
	}

	q1, q2, q3 := Quartiles(data)

	return &Summary{
		Count:    len(data),
		Mean:     Mean(data),
		Median:   Median(data),
		StdDev:   StdDev(data),
		Variance: Variance(data),
		Min:      Min(data),
		Max:      Max(data),
		Range:    Range(data),
		Q1:       q1,
		Q2:       q2,
		Q3:       q3,
		IQR:      IQR(data),
	}
}
