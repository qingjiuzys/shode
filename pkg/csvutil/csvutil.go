// Package csvutil 提供CSV处理工具
package csvutil

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Reader CSV读取器
type Reader struct {
	reader *csv.Reader
}

// NewReader 创建CSV读取器
func NewReader(r io.Reader) *Reader {
	return &Reader{
		reader: csv.NewReader(r),
	}
}

// NewReaderFromFile 从文件创建读取器
func NewReaderFromFile(filename string) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return &Reader{
		reader: csv.NewReader(file),
	}, nil
}

// ReadAll 读取所有记录
func (r *Reader) ReadAll() ([][]string, error) {
	return r.reader.ReadAll()
}

// Read 读取一条记录
func (r *Reader) Read() ([]string, error) {
	return r.reader.Read()
}

// ReadAsMap 读取为map
func (r *Reader) ReadAsMap() ([]map[string]string, error) {
	records, err := r.reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return []map[string]string{}, nil
	}

	headers := records[0]
	result := make([]map[string]string, 0, len(records)-1)

	for i := 1; i < len(records); i++ {
		row := records[i]
		m := make(map[string]string)

		for j, header := range headers {
			if j < len(row) {
				m[header] = row[j]
			} else {
				m[header] = ""
			}
		}

		result = append(result, m)
	}

	return result, nil
}

// ReadAsStruct 读取为结构体切片
func (r *Reader) ReadAsStruct(ptr any) error {
	records, err := r.reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	headers := records[0]
	rows := records[1:]

	// 使用反射处理结构体
	return mapToStruct(headers, rows, ptr)
}

func mapToStruct(headers []string, rows [][]string, ptr any) error {
	// 简化实现，实际应该使用反射
	// 这里返回JSON让调用者处理
	data := make([]map[string]string, 0, len(rows))

	for _, row := range rows {
		m := make(map[string]string)
		for j, header := range headers {
			if j < len(row) {
				m[header] = row[j]
			}
		}
		data = append(data, m)
	}

	jsonData, _ := json.Marshal(data)
	return json.Unmarshal(jsonData, ptr)
}

// SetComment 设置注释字符
func (r *Reader) SetComment(comment rune) {
	r.reader.Comment = comment
}

// SetFieldsPerRecord 设置每条记录的字段数
func (r *Reader) SetFieldsPerRecord(fields int) {
	r.reader.FieldsPerRecord = fields
}

// SetComma 设置字段分隔符为逗号
func (r *Reader) SetComma() {
	r.reader.Comma = ','
}

// SetTab 设置字段分隔符为制表符
func (r *Reader) SetTab() {
	r.reader.Comma = '\t'
}

// Writer CSV写入器
type Writer struct {
	writer *csv.Writer
}

// NewWriter 创建CSV写入器
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		writer: csv.NewWriter(w),
	}
}

// NewWriterToFile 创建文件写入器
func NewWriterToFile(filename string) (*Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return &Writer{
		writer: csv.NewWriter(file),
	}, nil
}

// Write 写入记录
func (w *Writer) Write(record []string) error {
	return w.writer.Write(record)
}

// WriteAll 写入所有记录
func (w *Writer) WriteAll(records [][]string) error {
	return w.writer.WriteAll(records)
}

// WriteHeader 写入表头
func (w *Writer) WriteHeader(headers []string) error {
	return w.writer.Write(headers)
}

// WriteMap 写入map数据
func (w *Writer) WriteMap(headers []string, data []map[string]string) error {
	if len(headers) == 0 {
		// 从第一条记录获取headers
		for k := range data[0] {
			headers = append(headers, k)
		}
	}

	// 写入表头
	if err := w.WriteHeader(headers); err != nil {
		return err
	}

	// 写入数据
	for _, m := range data {
		row := make([]string, len(headers))
		for i, header := range headers {
			row[i] = m[header]
		}

		if err := w.Write(row); err != nil {
			return err
		}
	}

	return nil
}

// Flush 刷新缓冲区
func (w *Writer) Flush() {
	w.writer.Flush()
}

// SetComma 设置分隔符为逗号
func (w *Writer) SetComma() {
	w.writer.Comma = ','
}

// SetTab 设置分隔符为制表符
func (w *Writer) SetTab() {
	w.writer.Comma = '\t'
}

// UseCRLF 使用CRLF作为行结束符
func (w *Writer) UseCRLF() {
	w.writer.UseCRLF = true
}

// ParseString 解析CSV字符串
func ParseString(data string) ([][]string, error) {
	reader := NewReader(strings.NewReader(data))
	return reader.ReadAll()
}

// ParseFile 解析CSV文件
func ParseFile(filename string) ([][]string, error) {
	reader, err := NewReaderFromFile(filename)
	if err != nil {
		return nil, err
	}

	return reader.ReadAll()
}

// ParseFileAsMap 解析CSV文件为map
func ParseFileAsMap(filename string) ([]map[string]string, error) {
	reader, err := NewReaderFromFile(filename)
	if err != nil {
		return nil, err
	}

	return reader.ReadAsMap()
}

// WriteString 写入CSV字符串
func WriteString(records [][]string) (string, error) {
	var buf bytes.Buffer
	writer := NewWriter(&buf)

	if err := writer.WriteAll(records); err != nil {
		return "", err
	}

	writer.Flush()
	return buf.String(), nil
}

// WriteToFile 写入CSV文件
func WriteToFile(filename string, records [][]string) error {
	writer, err := NewWriterToFile(filename)
	if err != nil {
		return err
	}

	defer writer.Flush()

	return writer.WriteAll(records)
}

// AppendTo 追加到CSV文件
func AppendTo(filename string, record []string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	return writer.Write(record)
}

// Merge 合并多个CSV文件
func Merge(filenames []string, outputFilename string) error {
	outFile, err := os.Create(outputFilename)
	if err != nil {
		return err
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()

	var headersWritten bool

	for _, filename := range filenames {
		reader, err := NewReaderFromFile(filename)
		if err != nil {
			return err
		}

		records, err := reader.ReadAll()
		if err != nil {
			return err
		}

		if len(records) == 0 {
			continue
		}

		if !headersWritten {
			// 写入表头
			if err := writer.WriteAll(records); err != nil {
				return err
			}
			headersWritten = true
		} else {
			// 跳过表头
			if len(records) > 1 {
				if err := writer.WriteAll(records[1:]); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Filter 过滤CSV行
func Filter(inputFilename, outputFilename string, filter func([]string) bool) error {
	reader, err := NewReaderFromFile(inputFilename)
	if err != nil {
		return err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// 过滤记录
	filtered := make([][]string, 0, len(records))
	filtered = append(filtered, records[0]) // 保留表头

	for i := 1; i < len(records); i++ {
		if filter(records[i]) {
			filtered = append(filtered, records[i])
		}
	}

	return WriteToFile(outputFilename, filtered)
}

// Transform 转换CSV数据
func Transform(inputFilename, outputFilename string, transform func([]string) []string) error {
	reader, err := NewReaderFromFile(inputFilename)
	if err != nil {
		return err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// 转换记录
	for i := range records {
		records[i] = transform(records[i])
	}

	return WriteToFile(outputFilename, records)
}

// Select 选择列
func Select(inputFilename, outputFilename string, columns []int) error {
	reader, err := NewReaderFromFile(inputFilename)
	if err != nil {
		return err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	// 选择列
	result := make([][]string, 0, len(records))
	for _, record := range records {
		row := make([]string, len(columns))
		for i, col := range columns {
			if col >= 0 && col < len(record) {
				row[i] = record[col]
			}
		}
		result = append(result, row)
	}

	return WriteToFile(outputFilename, result)
}

// ToJSON 转换为JSON
func ToJSON(csvFilename string) (string, error) {
	reader, err := NewReaderFromFile(csvFilename)
	if err != nil {
		return "", err
	}

	maps, err := reader.ReadAsMap()
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(maps, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// ToJSONFile 转换为JSON文件
func ToJSONFile(csvFilename, jsonFilename string) error {
	jsonStr, err := ToJSON(csvFilename)
	if err != nil {
		return err
	}

	return os.WriteFile(jsonFilename, []byte(jsonStr), 0644)
}

// FromJSON 从JSON创建CSV
func FromJSON(jsonFilename, csvFilename string) error {
	data, err := os.ReadFile(jsonFilename)
	if err != nil {
		return err
	}

	var maps []map[string]any
	if err := json.Unmarshal(data, &maps); err != nil {
		return err
	}

	if len(maps) == 0 {
		return errors.New("empty JSON data")
	}

	// 获取所有键作为表头
	var headers []string
	for k := range maps[0] {
		headers = append(headers, k)
	}

	// 转换为CSV记录
	records := make([][]string, 0, len(maps)+1)
	records = append(records, headers)

	for _, m := range maps {
		row := make([]string, len(headers))
		for i, header := range headers {
			if val, ok := m[header]; ok {
				row[i] = fmt.Sprintf("%v", val)
			}
		}
		records = append(records, row)
	}

	return WriteToFile(csvFilename, records)
}

// CountRows 统计行数
func CountRows(filename string) (int, error) {
	reader, err := NewReaderFromFile(filename)
	if err != nil {
		return 0, err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return 0, err
	}

	return len(records), nil
}

// CountColumns 统计列数
func CountColumns(filename string) (int, error) {
	reader, err := NewReaderFromFile(filename)
	if err != nil {
		return 0, err
	}

	record, err := reader.Read()
	if err != nil {
		return 0, err
	}

	return len(record), nil
}

// GetHeaders 获取表头
func GetHeaders(filename string) ([]string, error) {
	reader, err := NewReaderFromFile(filename)
	if err != nil {
		return nil, err
	}

	record, err := reader.Read()
	if err != nil {
		return nil, err
	}

	return record, nil
}

// GetColumn 获取列数据
func GetColumn(filename string, columnIndex int) ([]string, error) {
	reader, err := NewReaderFromFile(filename)
	if err != nil {
		return nil, err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return []string{}, nil
	}

	// 跳过表头
	result := make([]string, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		record := records[i]
		if columnIndex >= 0 && columnIndex < len(record) {
			result = append(result, record[columnIndex])
		}
	}

	return result, nil
}

// GetColumnByName 按列名获取数据
func GetColumnByName(filename, columnName string) ([]string, error) {
	reader, err := NewReaderFromFile(filename)
	if err != nil {
		return nil, err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return []string{}, nil
	}

	// 查找列索引
	headers := records[0]
	columnIndex := -1
	for i, header := range headers {
		if header == columnName {
			columnIndex = i
			break
		}
	}

	if columnIndex == -1 {
		return nil, fmt.Errorf("column not found: %s", columnName)
	}

	// 获取列数据
	result := make([]string, 0, len(records)-1)
	for i := 1; i < len(records); i++ {
		record := records[i]
		if columnIndex < len(record) {
			result = append(result, record[columnIndex])
		}
	}

	return result, nil
}

// Search 搜索包含指定文本的单元格
func Search(filename, searchText string, caseSensitive bool) ([]Cell, error) {
	reader, err := NewReaderFromFile(filename)
	if err != nil {
		return nil, err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if !caseSensitive {
		searchText = strings.ToLower(searchText)
	}

	var results []Cell

	for rowIdx, record := range records {
		for colIdx, cell := range record {
			cellText := cell
			if !caseSensitive {
				cellText = strings.ToLower(cell)
			}

			if strings.Contains(cellText, searchText) {
				results = append(results, Cell{
					Row:    rowIdx + 1,
					Column: colIdx + 1,
					Value:  cell,
				})
			}
		}
	}

	return results, nil
}

// Cell 单元格
type Cell struct {
	Row    int
	Column int
	Value  string
}

// SortByColumn 按列排序
func SortByColumn(inputFilename, outputFilename string, columnIndex int, ascending bool) error {
	reader, err := NewReaderFromFile(inputFilename)
	if err != nil {
		return err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) <= 1 {
		return WriteToFile(outputFilename, records)
	}

	// 保留表头，排序数据行
	headers := records[0]
	data := records[1:]

	// 简单冒泡排序
	for i := 0; i < len(data)-1; i++ {
		for j := 0; j < len(data)-i-1; j++ {
			row1 := data[j]
			row2 := data[j+1]

			// 比较指定列
			cmp := strings.Compare(row1[columnIndex], row2[columnIndex])
			if ascending {
				if cmp > 0 {
					data[j], data[j+1] = data[j+1], data[j]
				}
			} else {
				if cmp < 0 {
					data[j], data[j+1] = data[j+1], data[j]
				}
			}
		}
	}

	// 组合结果
	result := make([][]string, 0, len(records))
	result = append(result, headers)
	result = append(result, data...)

	return WriteToFile(outputFilename, result)
}

// GroupBy 分组
func GroupBy(inputFilename string, columnIndex int) (map[string][][]string, error) {
	reader, err := NewReaderFromFile(inputFilename)
	if err != nil {
		return nil, err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	groups := make(map[string][][]string)

	if len(records) <= 1 {
		return groups, nil
	}

	// 跳过表头
	for i := 1; i < len(records); i++ {
		record := records[i]
		if columnIndex >= 0 && columnIndex < len(record) {
			key := record[columnIndex]
			groups[key] = append(groups[key], record)
		}
	}

	return groups, nil
}

// Aggregate 聚合计算
func Aggregate(inputFilename string, columnIndex int, aggregateFunc func([]string) float64) (map[string]float64, error) {
	groups, err := GroupBy(inputFilename, columnIndex)
	if err != nil {
		return nil, err
	}

	result := make(map[string]float64)

	for key, rows := range groups {
		var values []string
		for _, row := range rows {
			if len(row) > columnIndex {
				values = append(values, row[columnIndex])
			}
		}

		result[key] = aggregateFunc(values)
	}

	return result, nil
}

// Sum 求和
func Sum(filename string, columnIndex int) (map[string]float64, error) {
	return Aggregate(filename, columnIndex, func(values []string) float64 {
		sum := 0.0
		for _, v := range values {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				sum += f
			}
		}
		return sum
	})
}

// Average 平均值
func Average(filename string, columnIndex int) (map[string]float64, error) {
	return Aggregate(filename, columnIndex, func(values []string) float64 {
		sum := 0.0
		count := 0

		for _, v := range values {
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				sum += f
				count++
			}
		}

		if count == 0 {
			return 0
		}

		return sum / float64(count)
	})
}

// Count 计数
func Count(filename string, columnIndex int) (map[string]int, error) {
	groups, err := GroupBy(filename, columnIndex)
	if err != nil {
		return nil, err
	}

	result := make(map[string]int)

	for key, rows := range groups {
		result[key] = len(rows)
	}

	return result, nil
}

// Validate 验证CSV格式
func Validate(filename string) error {
	reader, err := NewReaderFromFile(filename)
	if err != nil {
		return err
	}

	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) == 0 {
		return errors.New("empty CSV file")
	}

	// 检查每行的列数是否一致
	colCount := len(records[0])
	for i, record := range records {
		if len(record) != colCount {
			return fmt.Errorf("row %d has %d columns, expected %d", i+1, len(record), colCount)
		}
	}

	return nil
}

// Diff 比较两个CSV文件
func Diff(file1, file2 string) ([]string, error) {
	reader1, err := NewReaderFromFile(file1)
	if err != nil {
		return nil, err
	}

	reader2, err := NewReaderFromFile(file2)
	if err != nil {
		return nil, err
	}

	records1, err := reader1.ReadAll()
	if err != nil {
		return nil, err
	}

	records2, err := reader2.ReadAll()
	if err != nil {
		return nil, err
	}

	var diffs []string

	// 比较行数
	if len(records1) != len(records2) {
		diffs = append(diffs, fmt.Sprintf("Row count differs: %d vs %d", len(records1), len(records2)))
	}

	// 比较内容
	maxRows := len(records1)
	if len(records2) > maxRows {
		maxRows = len(records2)
	}

	for i := 0; i < maxRows; i++ {
		var row1, row2 []string

		if i < len(records1) {
			row1 = records1[i]
		}

		if i < len(records2) {
			row2 = records2[i]
		}

		if !equalSlices(row1, row2) {
			diffs = append(diffs, fmt.Sprintf("Row %d differs", i+1))
		}
	}

	return diffs, nil
}

func equalSlices(a, b []string) bool {
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
