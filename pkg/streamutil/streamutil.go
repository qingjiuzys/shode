// Package streamutil 提供流处理工具
package streamutil

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"sync"
)

// Stream 流接口
type Stream[T any] interface {
	// Next 获取下一个元素
	Next() bool
	// Value 获取当前元素
	Value() T
	// Close 关闭流
	Close() error
}

// SliceStream 切片流
type SliceStream[T any] struct {
	slice []T
	index int
}

// NewSliceStream 创建切片流
func NewSliceStream[T any](slice []T) *SliceStream[T] {
	return &SliceStream[T]{
		slice: slice,
		index: -1,
	}
}

// Next 获取下一个元素
func (s *SliceStream[T]) Next() bool {
	s.index++
	return s.index < len(s.slice)
}

// Value 获取当前元素
func (s *SliceStream[T]) Value() T {
	if s.index < 0 || s.index >= len(s.slice) {
		var zero T
		return zero
	}
	return s.slice[s.index]
}

// Close 关闭流
func (s *SliceStream[T]) Close() error {
	s.index = len(s.slice)
	return nil
}

// ReaderStream 读取器流
type ReaderStream struct {
	reader io.Reader
	buf    *bufio.Reader
	line   string
	err    error
}

// NewReaderStream 创建读取器流
func NewReaderStream(r io.Reader) *ReaderStream {
	return &ReaderStream{
		reader: r,
		buf:    bufio.NewReader(r),
	}
}

// Next 读取下一行
func (s *ReaderStream) Next() bool {
	s.line, s.err = s.buf.ReadString('\n')
	return s.err == nil && s.line != ""
}

// Value 获取当前行
func (s *ReaderStream) Value() string {
	return strings.TrimSuffix(s.line, "\n")
}

// Close 关闭流
func (s *ReaderStream) Close() error {
	if closer, ok := s.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// Err 获取错误
func (s *ReaderStream) Err() error {
	return s.err
}

// ChannelStream 通道流
type ChannelStream[T any] struct {
	ch    chan T
	value T
	ok    bool
}

// NewChannelStream 创建通道流
func NewChannelStream[T any](ch chan T) *ChannelStream[T] {
	return &ChannelStream[T]{
		ch: ch,
	}
}

// Next 获取下一个元素
func (s *ChannelStream[T]) Next() bool {
	s.value, s.ok = <-s.ch
	return s.ok
}

// Value 获取当前元素
func (s *ChannelStream[T]) Value() T {
	return s.value
}

// Close 关闭流
func (s *ChannelStream[T]) Close() error {
	close(s.ch)
	return nil
}

// Pipeline 管道
type Pipeline[T any, R any] struct {
	source Stream[T]
	fn     func(T) (R, error)
}

// NewPipeline 创建管道
func NewPipeline[T any, R any](source Stream[T], fn func(T) (R, error)) *Pipeline[T, R] {
	return &Pipeline[T, R]{
		source: source,
		fn:     fn,
	}
}

// Next 获取下一个处理后的元素
func (p *Pipeline[T, R]) Next() bool {
	return p.source.Next()
}

// Value 获取处理后的值
func (p *Pipeline[T, R]) Value() R {
	input := p.source.Value()
	output, _ := p.fn(input)
	return output
}

// Close 关闭管道
func (p *Pipeline[T, R]) Close() error {
	return p.source.Close()
}

// Filter 过滤流
func Filter[T any](stream Stream[T], predicate func(T) bool) Stream[T] {
	return &filterStream[T]{
		source:    stream,
		predicate: predicate,
	}
}

type filterStream[T any] struct {
	source    Stream[T]
	predicate func(T) bool
	current   T
	hasValue  bool
}

func (f *filterStream[T]) Next() bool {
	for f.source.Next() {
		value := f.source.Value()
		if f.predicate(value) {
			f.current = value
			f.hasValue = true
			return true
		}
	}
	f.hasValue = false
	return false
}

func (f *filterStream[T]) Value() T {
	return f.current
}

func (f *filterStream[T]) Close() error {
	return f.source.Close()
}

// Map 映射流
func Map[T any, R any](stream Stream[T], mapper func(T) R) Stream[R] {
	return &mapStream[T, R]{
		source: stream,
		mapper: mapper,
	}
}

type mapStream[T any, R any] struct {
	source Stream[T]
	mapper func(T) R
}

func (m *mapStream[T, R]) Next() bool {
	return m.source.Next()
}

func (m *mapStream[T, R]) Value() R {
	return m.mapper(m.source.Value())
}

func (m *mapStream[T, R]) Close() error {
	return m.source.Close()
}

// Take 获取前n个元素
func Take[T any](stream Stream[T], n int) Stream[T] {
	return &takeStream[T]{
		source: stream,
		n:      n,
		count:  0,
	}
}

type takeStream[T any] struct {
	source Stream[T]
	n      int
	count  int
}

func (t *takeStream[T]) Next() bool {
	if t.count >= t.n {
		return false
	}
	t.count++
	return t.source.Next()
}

func (t *takeStream[T]) Value() T {
	return t.source.Value()
}

func (t *takeStream[T]) Close() error {
	return t.source.Close()
}

// Skip 跳过前n个元素
func Skip[T any](stream Stream[T], n int) Stream[T] {
	return &skipStream[T]{
		source: stream,
		n:      n,
		skipped: false,
	}
}

type skipStream[T any] struct {
	source Stream[T]
	n      int
	skipped bool
	count  int
}

func (s *skipStream[T]) Next() bool {
	for {
		if !s.source.Next() {
			return false
		}
		if !s.skipped {
			s.count++
			if s.count >= s.n {
				s.skipped = true
			}
			continue
		}
		return true
	}
}

func (s *skipStream[T]) Value() T {
	return s.source.Value()
}

func (s *skipStream[T]) Close() error {
	return s.source.Close()
}

// Limit 限制流元素数量
func Limit[T any](stream Stream[T], n int) Stream[T] {
	return Take(stream, n)
}

// While 满足条件时继续
func While[T any](stream Stream[T], predicate func(T) bool) Stream[T] {
	return &whileStream[T]{
		source:    stream,
		predicate: predicate,
	}
}

type whileStream[T any] struct {
	source    Stream[T]
	predicate func(T) bool
	done      bool
}

func (w *whileStream[T]) Next() bool {
	if w.done {
		return false
	}
	if !w.source.Next() {
		return false
	}
	value := w.source.Value()
	if !w.predicate(value) {
		w.done = true
		return false
	}
	return true
}

func (w *whileStream[T]) Value() T {
	return w.source.Value()
}

func (w *whileStream[T]) Close() error {
	return w.source.Close()
}

// Collect 收集流元素到切片
func Collect[T any](stream Stream[T]) ([]T, error) {
	defer stream.Close()

	var result []T
	for stream.Next() {
		result = append(result, stream.Value())
	}
	return result, nil
}

// CollectMap 收集流元素到map
func CollectMap[K comparable, V any](stream Stream[V], keyFunc func(V) K) (map[K]V, error) {
	defer stream.Close()

	result := make(map[K]V)
	for stream.Next() {
		value := stream.Value()
		key := keyFunc(value)
		result[key] = value
	}
	return result, nil
}

// ForEach 遍历流
func ForEach[T any](stream Stream[T], fn func(T)) error {
	defer stream.Close()

	for stream.Next() {
		fn(stream.Value())
	}
	return nil
}

// Reduce 归约流
func Reduce[T any, R any](stream Stream[T], initial R, fn func(R, T) R) (R, error) {
	defer stream.Close()

	result := initial
	for stream.Next() {
		result = fn(result, stream.Value())
	}
	return result, nil
}

// Count 统计流元素数量
func Count[T any](stream Stream[T]) (int, error) {
	count := 0
	err := ForEach(stream, func(T) {
		count++
	})
	return count, err
}

// Any 是否有元素满足条件
func Any[T any](stream Stream[T], predicate func(T) bool) (bool, error) {
	defer stream.Close()

	for stream.Next() {
		if predicate(stream.Value()) {
			return true, nil
		}
	}
	return false, nil
}

// All 是否所有元素满足条件
func All[T any](stream Stream[T], predicate func(T) bool) (bool, error) {
	defer stream.Close()

	for stream.Next() {
		if !predicate(stream.Value()) {
			return false, nil
		}
	}
	return true, nil
}

// First 获取第一个元素
func First[T any](stream Stream[T]) (T, error) {
	defer stream.Close()

	if stream.Next() {
		return stream.Value(), nil
	}
	var zero T
	return zero, io.EOF
}

// Find 查找满足条件的元素
func Find[T any](stream Stream[T], predicate func(T) bool) (T, error) {
	defer stream.Close()

	for stream.Next() {
		value := stream.Value()
		if predicate(value) {
			return value, nil
		}
	}
	var zero T
	return zero, io.EOF
}

// Chunk 分块流
func Chunk[T any](stream Stream[T], size int) Stream[[]T] {
	return &chunkStream[T]{
		source: stream,
		size:   size,
	}
}

type chunkStream[T any] struct {
	source Stream[T]
	size   int
	buffer []T
}

func (c *chunkStream[T]) Next() bool {
	c.buffer = nil

	for i := 0; i < c.size; i++ {
		if !c.source.Next() {
			return len(c.buffer) > 0
		}
		c.buffer = append(c.buffer, c.source.Value())
	}
	return true
}

func (c *chunkStream[T]) Value() []T {
	return c.buffer
}

func (c *chunkStream[T]) Close() error {
	return c.source.Close()
}

// Flatten 扁平化流
func Flatten[T any](stream Stream[[]T]) Stream[T] {
	return &flattenStream[T]{
		source: stream,
		buffer: nil,
		index:  0,
	}
}

type flattenStream[T any] struct {
	source Stream[[]T]
	buffer []T
	index  int
}

func (f *flattenStream[T]) Next() bool {
	for {
		if f.buffer != nil && f.index < len(f.buffer) {
			f.index++
			return true
		}

		if !f.source.Next() {
			return false
		}

		f.buffer = f.source.Value()
		f.index = 0
	}
}

func (f *flattenStream[T]) Value() T {
	if f.buffer != nil && f.index < len(f.buffer) {
		return f.buffer[f.index]
	}
	var zero T
	return zero
}

func (f *flattenStream[T]) Close() error {
	return f.source.Close()
}

// Merge 合并多个流
func Merge[T any](streams ...Stream[T]) Stream[T] {
	return &mergeStream[T]{
		streams: streams,
		index:   0,
	}
}

type mergeStream[T any] struct {
	streams []Stream[T]
	index   int
}

func (m *mergeStream[T]) Next() bool {
	for m.index < len(m.streams) {
		if m.streams[m.index].Next() {
			return true
		}
		m.streams[m.index].Close()
		m.index++
	}
	return false
}

func (m *mergeStream[T]) Value() T {
	if m.index < len(m.streams) {
		return m.streams[m.index].Value()
	}
	var zero T
	return zero
}

func (m *mergeStream[T]) Close() error {
	for _, stream := range m.streams {
		stream.Close()
	}
	return nil
}

// Zip 拉链流
func Zip[T any](streams ...Stream[T]) Stream[[]T] {
	return &zipStream[T]{
		streams: streams,
	}
}

type zipStream[T any] struct {
	streams []Stream[T]
}

func (z *zipStream[T]) Next() bool {
	for _, stream := range z.streams {
		if !stream.Next() {
			return false
		}
	}
	return true
}

func (z *zipStream[T]) Value() []T {
	result := make([]T, len(z.streams))
	for i, stream := range z.streams {
		result[i] = stream.Value()
	}
	return result
}

func (z *zipStream[T]) Close() error {
	for _, stream := range z.streams {
		stream.Close()
	}
	return nil
}

// Distinct 去重流
func Distinct[T comparable](stream Stream[T]) Stream[T] {
	return &distinctStream[T]{
		source: stream,
		seen:   make(map[T]struct{}),
	}
}

type distinctStream[T comparable] struct {
	source Stream[T]
	seen   map[T]struct{}
	current T
}

func (d *distinctStream[T]) Next() bool {
	for d.source.Next() {
		value := d.source.Value()
		if _, exists := d.seen[value]; !exists {
			d.seen[value] = struct{}{}
			d.current = value
			return true
		}
	}
	return false
}

func (d *distinctStream[T]) Value() T {
	return d.current
}

func (d *distinctStream[T]) Close() error {
	return d.source.Close()
}

// Peek 偷看每个元素
func Peek[T any](stream Stream[T], fn func(T)) Stream[T] {
	return &peekStream[T]{
		source: stream,
		fn:     fn,
	}
}

type peekStream[T any] struct {
	source Stream[T]
	fn     func(T)
}

func (p *peekStream[T]) Next() bool {
	return p.source.Next()
}

func (p *peekStream[T]) Value() T {
	value := p.source.Value()
	p.fn(value)
	return value
}

func (p *peekStream[T]) Close() error {
	return p.source.Close()
}

// Buffer 缓冲流
func Buffer[T any](stream Stream[T], size int) Stream[T] {
	return &bufferStream[T]{
		source: stream,
		buffer: make([]T, 0, size),
	}
}

type bufferStream[T any] struct {
	source Stream[T]
	buffer []T
	mu     sync.Mutex
}

func (b *bufferStream[T]) Next() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.buffer) > 0 {
		b.buffer = b.buffer[1:]
		return true
	}

	return b.source.Next()
}

func (b *bufferStream[T]) Value() T {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.buffer) > 0 {
		return b.buffer[0]
	}

	return b.source.Value()
}

func (b *bufferStream[T]) Close() error {
	return b.source.Close()
}

// Split 分割流
func Split(stream Stream[string], sep string) Stream[string] {
	return &splitStream{
		source: stream,
		sep:    sep,
		parts:  nil,
		index:  0,
	}
}

type splitStream struct {
	source Stream[string]
	sep    string
	parts  []string
	index  int
}

func (s *splitStream) Next() bool {
	for {
		if s.parts != nil && s.index < len(s.parts) {
			s.index++
			return true
		}

		if !s.source.Next() {
			return false
		}

		s.parts = strings.Split(s.source.Value(), s.sep)
		s.index = 0
	}
}

func (s *splitStream) Value() string {
	if s.parts != nil && s.index < len(s.parts) {
		return s.parts[s.index]
	}
	return ""
}

func (s *splitStream) Close() error {
	return s.source.Close()
}

// Join 连接流
func Join[T any](stream Stream[T], sep string) Stream[string] {
	return Map(stream, func(v T) string {
		return fmt.Sprintf("%v", v)
	})
}

// Lines 按行分割流
func Lines(r io.Reader) Stream[string] {
	return NewReaderStream(r)
}

// Strings 从字符串创建流
func Strings(s string) Stream[string] {
	return Lines(strings.NewReader(s))
}

// Bytes 从字节创建流
func Bytes(data []byte) Stream[byte] {
	slice := make([]byte, len(data))
	copy(slice, data)
	return NewSliceStream(slice)
}

// ReaderToStream 将io.Reader转为字符串流
func ReaderToStream(r io.Reader) Stream[string] {
	return NewReaderStream(r)
}

// StreamToWriter 将流写入io.Writer
func StreamToWriter[T any](stream Stream[T], writer io.Writer, format func(T) string) (int, error) {
	defer stream.Close()

	count := 0
	for stream.Next() {
		value := stream.Value()
		data := []byte(fmt.Sprintf("%v\n", value))
		if format != nil {
			data = []byte(format(value))
		}
		n, err := writer.Write(data)
		if err != nil {
			return count, err
		}
		count += n
	}
	return count, nil
}

// StreamToBytes 将流转为字节
func StreamToBytes[T any](stream Stream[T], format func(T) string) ([]byte, error) {
	var buf bytes.Buffer
	_, err := StreamToWriter(stream, &buf, format)
	return buf.Bytes(), err
}

// Tee 分支流
func Tee[T any](stream Stream[T], fn func(T)) Stream[T] {
	return Peek(stream, fn)
}

// Batch 批处理流
func Batch[T any](stream Stream[T], size int, fn func([]T)) error {
	defer stream.Close()

	buffer := make([]T, 0, size)
	for stream.Next() {
		buffer = append(buffer, stream.Value())
		if len(buffer) >= size {
			fn(buffer)
			buffer = make([]T, 0, size)
		}
	}

	if len(buffer) > 0 {
		fn(buffer)
	}

	return nil
}

// Parallel 并行处理流
func Parallel[T any](stream Stream[T], workers int, fn func(T)) error {
	defer stream.Close()

	wg := &sync.WaitGroup{}
	ch := make(chan T, workers*2)

	// 启动worker
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range ch {
				fn(item)
			}
		}()
	}

	// 发送数据
	for stream.Next() {
		ch <- stream.Value()
	}
	close(ch)

	wg.Wait()
	return nil
}
