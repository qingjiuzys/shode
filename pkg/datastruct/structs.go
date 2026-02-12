// Package datastruct 提供常用数据结构
package datastruct

// Ordered 可排序类型约束
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 | ~string
}

// Stack 栈结构
type Stack[T any] struct {
	items []T
}

// NewStack 创建栈
func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		items: make([]T, 0),
	}
}

// Push 压入元素
func (s *Stack[T]) Push(item T) {
	s.items = append(s.items, item)
}

// Pop 弹出元素
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}

	index := len(s.items) - 1
	item := s.items[index]
	s.items = s.items[:index]
	return item, true
}

// Peek 查看栈顶元素
func (s *Stack[T]) Peek() (T, bool) {
	if len(s.items) == 0 {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

// IsEmpty 检查是否为空
func (s *Stack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

// Size 获取大小
func (s *Stack[T]) Size() int {
	return len(s.items)
}

// Clear 清空栈
func (s *Stack[T]) Clear() {
	s.items = make([]T, 0)
}

// ToSlice 转换为切片
func (s *Stack[T]) ToSlice() []T {
	return s.items
}

// Queue 队列结构
type Queue[T any] struct {
	items []T
	head  int
	tail  int
	size  int
}

// NewQueue 创建队列
func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		items: make([]T, 16),
	}
}

// Enqueue 入队
func (q *Queue[T]) Enqueue(item T) {
	if q.size == len(q.items) {
		// 扩容
		newItems := make([]T, len(q.items)*2)
		copy(newItems, q.items[q.head:])
		copy(newItems[len(q.items)-q.head:], q.items[:q.tail])
		q.items = newItems
		q.head = 0
		q.tail = q.size
	}

	q.items[q.tail] = item
	q.tail = (q.tail + 1) % len(q.items)
	q.size++
}

// Dequeue 出队
func (q *Queue[T]) Dequeue() (T, bool) {
	if q.size == 0 {
		var zero T
		return zero, false
	}

	item := q.items[q.head]
	q.head = (q.head + 1) % len(q.items)
	q.size--

	// 缩容
	if q.size > 0 && q.size == len(q.items)/4 {
		newItems := make([]T, len(q.items)/2)
		copy(newItems, q.items[q.head:])
		copy(newItems[len(q.items)-q.head:], q.items[:q.tail])
		q.items = newItems
		q.head = 0
		q.tail = q.size
	}

	return item, true
}

// Peek 查看队首元素
func (q *Queue[T]) Peek() (T, bool) {
	if q.size == 0 {
		var zero T
		return zero, false
	}
	return q.items[q.head], true
}

// IsEmpty 检查是否为空
func (q *Queue[T]) IsEmpty() bool {
	return q.size == 0
}

// Size 获取大小
func (q *Queue[T]) Size() int {
	return q.size
}

// Clear 清空队列
func (q *Queue[T]) Clear() {
	q.items = make([]T, 16)
	q.head = 0
	q.tail = 0
	q.size = 0
}

// ToSlice 转换为切片
func (q *Queue[T]) ToSlice() []T {
	result := make([]T, 0, q.size)
	for i := 0; i < q.size; i++ {
		result = append(result, q.items[(q.head+i)%len(q.items)])
	}
	return result
}

// LinkedList 链表结构
type LinkedList[T any] struct {
	head *Node[T]
	tail *Node[T]
	size int
}

// Node 链表节点
type Node[T any] struct {
	Value T
	Next  *Node[T]
	Prev  *Node[T]
}

// NewLinkedList 创建链表
func NewLinkedList[T any]() *LinkedList[T] {
	return &LinkedList[T]{}
}

// PushFront 在头部插入
func (l *LinkedList[T]) PushFront(value T) {
	node := &Node[T]{Value: value}

	if l.head == nil {
		l.head = node
		l.tail = node
	} else {
		node.Next = l.head
		l.head.Prev = node
		l.head = node
	}
	l.size++
}

// PushBack 在尾部插入
func (l *LinkedList[T]) PushBack(value T) {
	node := &Node[T]{Value: value}

	if l.tail == nil {
		l.head = node
		l.tail = node
	} else {
		node.Prev = l.tail
		l.tail.Next = node
		l.tail = node
	}
	l.size++
}

// PopFront 删除头部节点
func (l *LinkedList[T]) PopFront() (T, bool) {
	if l.head == nil {
		var zero T
		return zero, false
	}

	value := l.head.Value
	l.head = l.head.Next

	if l.head != nil {
		l.head.Prev = nil
	} else {
		l.tail = nil
	}
	l.size--

	return value, true
}

// PopBack 删除尾部节点
func (l *LinkedList[T]) PopBack() (T, bool) {
	if l.tail == nil {
		var zero T
		return zero, false
	}

	value := l.tail.Value
	l.tail = l.tail.Prev

	if l.tail != nil {
		l.tail.Next = nil
	} else {
		l.head = nil
	}
	l.size--

	return value, true
}

// Front 获取头部元素
func (l *LinkedList[T]) Front() (T, bool) {
	if l.head == nil {
		var zero T
		return zero, false
	}
	return l.head.Value, true
}

// Back 获取尾部元素
func (l *LinkedList[T]) Back() (T, bool) {
	if l.tail == nil {
		var zero T
		return zero, false
	}
	return l.tail.Value, true
}

// IsEmpty 检查是否为空
func (l *LinkedList[T]) IsEmpty() bool {
	return l.size == 0
}

// Size 获取大小
func (l *LinkedList[T]) Size() int {
	return l.size
}

// Clear 清空链表
func (l *LinkedList[T]) Clear() {
	l.head = nil
	l.tail = nil
	l.size = 0
}

// ToSlice 转换为切片
func (l *LinkedList[T]) ToSlice() []T {
	result := make([]T, 0, l.size)
	for node := l.head; node != nil; node = node.Next {
		result = append(result, node.Value)
	}
	return result
}

// Find 查找元素
func (l *LinkedList[T]) Find(value T) *Node[T] {
	for node := l.head; node != nil; node = node.Next {
		// 使用==比较（适用于可比较类型）
		// 对于接口类型，可能需要其他比较方式
		if any(node.Value) == any(value) {
			return node
		}
	}
	return nil
}

// Remove 删除节点
func (l *LinkedList[T]) Remove(node *Node[T]) bool {
	if node == nil {
		return false
	}

	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		l.head = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		l.tail = node.Prev
	}

	l.size--
	return true
}

// BinarySearchTree 二叉搜索树
type BinarySearchTree[T Ordered] struct {
	root *TreeNode[T]
	size int
}

// TreeNode 树节点
type TreeNode[T Ordered] struct {
	Value T
	Left  *TreeNode[T]
	Right *TreeNode[T]
}

// NewBinarySearchTree 创建二叉搜索树
func NewBinarySearchTree[T Ordered]() *BinarySearchTree[T] {
	return &BinarySearchTree[T]{}
}

// Insert 插入值
func (bst *BinarySearchTree[T]) Insert(value T) {
	bst.root = bst.insert(bst.root, value)
	bst.size++
}

func (bst *BinarySearchTree[T]) insert(node *TreeNode[T], value T) *TreeNode[T] {
	if node == nil {
		return &TreeNode[T]{Value: value}
	}

	if value < node.Value {
		node.Left = bst.insert(node.Left, value)
	} else if value > node.Value {
		node.Right = bst.insert(node.Right, value)
	}

	return node
}

// Search 搜索值
func (bst *BinarySearchTree[T]) Search(value T) bool {
	return bst.search(bst.root, value)
}

func (bst *BinarySearchTree[T]) search(node *TreeNode[T], value T) bool {
	if node == nil {
		return false
	}

	if value == node.Value {
		return true
	} else if value < node.Value {
		return bst.search(node.Left, value)
	} else {
		return bst.search(node.Right, value)
	}
}

// Delete 删除值
func (bst *BinarySearchTree[T]) Delete(value T) {
	if bst.Search(value) {
		bst.root = bst.delete(bst.root, value)
		bst.size--
	}
}

func (bst *BinarySearchTree[T]) delete(node *TreeNode[T], value T) *TreeNode[T] {
	if node == nil {
		return nil
	}

	if value < node.Value {
		node.Left = bst.delete(node.Left, value)
	} else if value > node.Value {
		node.Right = bst.delete(node.Right, value)
	} else {
		// 找到节点，执行删除

		// 情况1: 没有左子节点
		if node.Left == nil {
			return node.Right
		}

		// 情况2: 没有右子节点
		if node.Right == nil {
			return node.Left
		}

		// 情况3: 有两个子节点
		// 找到右子树的最小值
		minNode := bst.findMin(node.Right)
		node.Value = minNode.Value
		node.Right = bst.delete(node.Right, minNode.Value)
	}

	return node
}

func (bst *BinarySearchTree[T]) findMin(node *TreeNode[T]) *TreeNode[T] {
	for node.Left != nil {
		node = node.Left
	}
	return node
}

// InOrder 中序遍历
func (bst *BinarySearchTree[T]) InOrder() []T {
	result := make([]T, 0, bst.size)
	bst.inOrder(bst.root, &result)
	return result
}

func (bst *BinarySearchTree[T]) inOrder(node *TreeNode[T], result *[]T) {
	if node == nil {
		return
	}
	bst.inOrder(node.Left, result)
	*result = append(*result, node.Value)
	bst.inOrder(node.Right, result)
}

// PreOrder 前序遍历
func (bst *BinarySearchTree[T]) PreOrder() []T {
	result := make([]T, 0, bst.size)
	bst.preOrder(bst.root, &result)
	return result
}

func (bst *BinarySearchTree[T]) preOrder(node *TreeNode[T], result *[]T) {
	if node == nil {
		return
	}
	*result = append(*result, node.Value)
	bst.preOrder(node.Left, result)
	bst.preOrder(node.Right, result)
}

// PostOrder 后序遍历
func (bst *BinarySearchTree[T]) PostOrder() []T {
	result := make([]T, 0, bst.size)
	bst.postOrder(bst.root, &result)
	return result
}

func (bst *BinarySearchTree[T]) postOrder(node *TreeNode[T], result *[]T) {
	if node == nil {
		return
	}
	bst.postOrder(node.Left, result)
	bst.postOrder(node.Right, result)
	*result = append(*result, node.Value)
}

// IsEmpty 检查是否为空
func (bst *BinarySearchTree[T]) IsEmpty() bool {
	return bst.size == 0
}

// Size 获取大小
func (bst *BinarySearchTree[T]) Size() int {
	return bst.size
}

// Clear 清空树
func (bst *BinarySearchTree[T]) Clear() {
	bst.root = nil
	bst.size = 0
}

// MinHeap 最小堆
type MinHeap[T Ordered] struct {
	items []T
}

// NewMinHeap 创建最小堆
func NewMinHeap[T Ordered]() *MinHeap[T] {
	return &MinHeap[T]{
		items: make([]T, 0),
	}
}

// Push 插入元素
func (h *MinHeap[T]) Push(item T) {
	h.items = append(h.items, item)
	h.heapifyUp(len(h.items) - 1)
}

// Pop 弹出最小元素
func (h *MinHeap[T]) Pop() (T, bool) {
	if len(h.items) == 0 {
		var zero T
		return zero, false
	}

	min := h.items[0]
	last := h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]

	if len(h.items) > 0 {
		h.items[0] = last
		h.heapifyDown(0)
	}

	return min, true
}

// Peek 查看最小元素
func (h *MinHeap[T]) Peek() (T, bool) {
	if len(h.items) == 0 {
		var zero T
		return zero, false
	}
	return h.items[0], true
}

func (h *MinHeap[T]) heapifyUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		if h.items[index] >= h.items[parent] {
			break
		}
		h.items[index], h.items[parent] = h.items[parent], h.items[index]
		index = parent
	}
}

func (h *MinHeap[T]) heapifyDown(index int) {
	for {
		left := 2*index + 1
		right := 2*index + 2
		smallest := index

		if left < len(h.items) && h.items[left] < h.items[smallest] {
			smallest = left
		}

		if right < len(h.items) && h.items[right] < h.items[smallest] {
			smallest = right
		}

		if smallest == index {
			break
		}

		h.items[index], h.items[smallest] = h.items[smallest], h.items[index]
		index = smallest
	}
}

// Size 获取大小
func (h *MinHeap[T]) Size() int {
	return len(h.items)
}

// IsEmpty 检查是否为空
func (h *MinHeap[T]) IsEmpty() bool {
	return len(h.items) == 0
}

// ToSlice 转换为切片
func (h *MinHeap[T]) ToSlice() []T {
	return h.items
}

// MaxHeap 最大堆
type MaxHeap[T Ordered] struct {
	items []T
}

// NewMaxHeap 创建最大堆
func NewMaxHeap[T Ordered]() *MaxHeap[T] {
	return &MaxHeap[T]{
		items: make([]T, 0),
	}
}

// Push 插入元素
func (h *MaxHeap[T]) Push(item T) {
	h.items = append(h.items, item)
	h.heapifyUp(len(h.items) - 1)
}

// Pop 弹出最大元素
func (h *MaxHeap[T]) Pop() (T, bool) {
	if len(h.items) == 0 {
		var zero T
		return zero, false
	}

	max := h.items[0]
	last := h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]

	if len(h.items) > 0 {
		h.items[0] = last
		h.heapifyDown(0)
	}

	return max, true
}

// Peek 查看最大元素
func (h *MaxHeap[T]) Peek() (T, bool) {
	if len(h.items) == 0 {
		var zero T
		return zero, false
	}
	return h.items[0], true
}

func (h *MaxHeap[T]) heapifyUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		if h.items[index] <= h.items[parent] {
			break
		}
		h.items[index], h.items[parent] = h.items[parent], h.items[index]
		index = parent
	}
}

func (h *MaxHeap[T]) heapifyDown(index int) {
	for {
		left := 2*index + 1
		right := 2*index + 2
		largest := index

		if left < len(h.items) && h.items[left] > h.items[largest] {
			largest = left
		}

		if right < len(h.items) && h.items[right] > h.items[largest] {
			largest = right
		}

		if largest == index {
			break
		}

		h.items[index], h.items[largest] = h.items[largest], h.items[index]
		index = largest
	}
}

// Size 获取大小
func (h *MaxHeap[T]) Size() int {
	return len(h.items)
}

// IsEmpty 检查是否为空
func (h *MaxHeap[T]) IsEmpty() bool {
	return len(h.items) == 0
}

// ToSlice 转换为切片
func (h *MaxHeap[T]) ToSlice() []T {
	return h.items
}

// Trie 前缀树
type Trie struct {
	root *TrieNode
	size int
}

// TrieNode 前缀树节点
type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	value    string
}

// NewTrie 创建前缀树
func NewTrie() *Trie {
	return &Trie{
		root: &TrieNode{
			children: make(map[rune]*TrieNode),
		},
	}
}

// Insert 插入单词
func (t *Trie) Insert(word string) {
	node := t.root
	for _, ch := range word {
		if node.children[ch] == nil {
			node.children[ch] = &TrieNode{
				children: make(map[rune]*TrieNode),
			}
		}
		node = node.children[ch]
	}
	node.isEnd = true
	node.value = word
	t.size++
}

// Search 搜索单词
func (t *Trie) Search(word string) bool {
	node := t.root
	for _, ch := range word {
		if node.children[ch] == nil {
			return false
		}
		node = node.children[ch]
	}
	return node.isEnd
}

// StartsWith 检查前缀
func (t *Trie) StartsWith(prefix string) bool {
	node := t.root
	for _, ch := range prefix {
		if node.children[ch] == nil {
			return false
		}
		node = node.children[ch]
	}
	return true
}

// Delete 删除单词
func (t *Trie) Delete(word string) {
	if t.Search(word) {
		t.delete(t.root, word, 0)
		t.size--
	}
}

func (t *Trie) delete(node *TrieNode, word string, depth int) bool {
	if depth == len(word) {
		if !node.isEnd {
			return false
		}
		node.isEnd = false
		return len(node.children) == 0
	}

	ch := rune(word[depth])
	child, exists := node.children[ch]
	if !exists {
		return false
	}

	shouldDelete := t.delete(child, word, depth+1)

	if shouldDelete {
		delete(node.children, ch)
		return len(node.children) == 0
	}

	return false
}

// Size 获取大小
func (t *Trie) Size() int {
	return t.size
}

// IsEmpty 检查是否为空
func (t *Trie) IsEmpty() bool {
	return t.size == 0
}

// Clear 清空前缀树
func (t *Trie) Clear() {
	t.root = &TrieNode{
		children: make(map[rune]*TrieNode),
	}
	t.size = 0
}

// AutoComplete 自动补全
func (t *Trie) AutoComplete(prefix string) []string {
	node := t.root
	for _, ch := range prefix {
		if node.children[ch] == nil {
			return []string{}
		}
		node = node.children[ch]
	}

	results := make([]string, 0)
	t.findAll(node, &results)
	return results
}

func (t *Trie) findAll(node *TrieNode, results *[]string) {
	if node.isEnd {
		*results = append(*results, node.value)
	}

	for _, child := range node.children {
		t.findAll(child, results)
	}
}

// LRUCache LRU缓存
type LRUCache[K comparable, V any] struct {
	capacity int
	cache    map[K]*lruNode[K, V]
	head     *lruNode[K, V]
	tail     *lruNode[K, V]
}

type lruNode[K comparable, V any] struct {
	key   K
	value V
	prev  *lruNode[K, V]
	next  *lruNode[K, V]
}

// NewLRUCache 创建LRU缓存
func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	cache := &LRUCache[K, V]{
		capacity: capacity,
		cache:    make(map[K]*lruNode[K, V]),
	}
	cache.head = &lruNode[K, V]{}
	cache.tail = &lruNode[K, V]{}
	cache.head.next = cache.tail
	cache.tail.prev = cache.head
	return cache
}

// Put 存入键值
func (c *LRUCache[K, V]) Put(key K, value V) {
	if node, exists := c.cache[key]; exists {
		node.value = value
		c.removeNode(node)
		c.addToFront(node)
		return
	}

	node := &lruNode[K, V]{key: key, value: value}
	c.cache[key] = node
	c.addToFront(node)

	if len(c.cache) > c.capacity {
		oldest := c.tail.prev
		c.removeNode(oldest)
		delete(c.cache, oldest.key)
	}
}

// Get 获取值
func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	node, exists := c.cache[key]
	if !exists {
		var zero V
		return zero, false
	}

	c.removeNode(node)
	c.addToFront(node)
	return node.value, true
}

// Delete 删除键值
func (c *LRUCache[K, V]) Delete(key K) {
	if node, exists := c.cache[key]; exists {
		c.removeNode(node)
		delete(c.cache, key)
	}
}

func (c *LRUCache[K, V]) addToFront(node *lruNode[K, V]) {
	node.prev = c.head
	node.next = c.head.next
	c.head.next.prev = node
	c.head.next = node
}

func (c *LRUCache[K, V]) removeNode(node *lruNode[K, V]) {
	node.prev.next = node.next
	node.next.prev = node.prev
}

// Size 获取大小
func (c *LRUCache[K, V]) Size() int {
	return len(c.cache)
}

// IsEmpty 检查是否为空
func (c *LRUCache[K, V]) IsEmpty() bool {
	return len(c.cache) == 0
}

// Clear 清空缓存
func (c *LRUCache[K, V]) Clear() {
	c.cache = make(map[K]*lruNode[K, V])
	c.head.next = c.tail
	c.tail.prev = c.head
}

// Keys 获取所有键
func (c *LRUCache[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.cache))
	for k := range c.cache {
		keys = append(keys, k)
	}
	return keys
}
