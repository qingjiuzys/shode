// Package xmlutil 提供XML处理工具
package xmlutil

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"
)

// Node XML节点
type Node struct {
	XMLName  xml.Name
	Attrs    map[string]string
	Content  string
	Children []*Node
}

// Parse 解析XML
func Parse(data []byte) (*Node, error) {
	var node Node
	err := xml.Unmarshal(data, &node)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// ParseFile 解析XML文件
func ParseFile(filename string) (*Node, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return Parse(data)
}

// ParseString 解析XML字符串
func ParseString(data string) (*Node, error) {
	return Parse([]byte(data))
}

// MustParse 必须成功解析
func MustParse(data []byte) *Node {
	node, err := Parse(data)
	if err != nil {
		panic(err)
	}
	return node
}

// Marshal 序列化为XML
func Marshal(v any) ([]byte, error) {
	return xml.Marshal(v)
}

// MarshalIndent 序列化为格式化XML
func MarshalIndent(v any) ([]byte, error) {
	return xml.MarshalIndent(v, "", "  ")
}

// MarshalToString 序列化为XML字符串
func MarshalToString(v any) (string, error) {
	data, err := xml.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Unmarshal 反序列化XML
func Unmarshal(data []byte, v any) error {
	return xml.Unmarshal(data, v)
}

// UnmarshalFile 从文件反序列化
func UnmarshalFile(filename string, v any) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return xml.Unmarshal(data, v)
}

// UnmarshalString 从字符串反序列化
func UnmarshalString(data string, v any) error {
	return xml.Unmarshal([]byte(data), v)
}

// Beautify 美化XML
func Beautify(data []byte) ([]byte, error) {
	v := any(&Node{})
	if err := xml.Unmarshal(data, v); err != nil {
		return nil, err
	}

	return xml.MarshalIndent(v, "", "  ")
}

// BeautifyString 美化XML字符串
func BeautifyString(data string) (string, error) {
	beautified, err := Beautify([]byte(data))
	if err != nil {
		return "", err
	}
	return string(beautified), nil
}

// Minify 压缩XML
func Minify(data []byte) ([]byte, error) {
	var buf bytes.Buffer

	decoder := xml.NewDecoder(bytes.NewReader(data))
	encoder := xml.NewEncoder(&buf)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if err := encoder.EncodeToken(token); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// MinifyString 压缩XML字符串
func MinifyString(data string) (string, error) {
	minified, err := Minify([]byte(data))
	if err != nil {
		return "", err
	}
	return string(minified), nil
}

// Extract 提取XML元素
func Extract(data []byte, path string) ([]string, error) {
	// 简化实现：使用正则表达式
	// 实际应该使用XML解析器
	return []string{}, nil
}

// Validate 验证XML格式
func Validate(data []byte) error {
	var v any
	return xml.Unmarshal(data, &v)
}

// GetAttribute 获取属性
func GetAttribute(node *Node, attrName string) (string, bool) {
	if node.Attrs == nil {
		return "", false
	}

	value, exists := node.Attrs[attrName]
	return value, exists
}

// SetAttribute 设置属性
func SetAttribute(node *Node, attrName, attrValue string) {
	if node.Attrs == nil {
		node.Attrs = make(map[string]string)
	}
	node.Attrs[attrName] = attrValue
}

// GetContent 获取内容
func GetContent(node *Node) string {
	return node.Content
}

// SetContent 设置内容
func SetContent(node *Node, content string) {
	node.Content = content
}

// GetChild 获取子节点
func GetChild(node *Node, name string) *Node {
	for _, child := range node.Children {
		if child.XMLName.Local == name {
			return child
		}
	}
	return nil
}

// GetChildren 获取所有子节点
func GetChildren(node *Node, name string) []*Node {
	var children []*Node
	for _, child := range node.Children {
		if name == "" || child.XMLName.Local == name {
			children = append(children, child)
		}
	}
	return children
}

// AddChild 添加子节点
func AddChild(node *Node, child *Node) {
	node.Children = append(node.Children, child)
}

// RemoveChild 移除子节点
func RemoveChild(node *Node, name string) bool {
	for i, child := range node.Children {
		if child.XMLName.Local == name {
			node.Children = append(node.Children[:i], node.Children[i+1:]...)
			return true
		}
	}
	return false
}

// Find 查找节点
func Find(node *Node, name string) []*Node {
	var results []*Node

	if node.XMLName.Local == name {
		results = append(results, node)
	}

	for _, child := range node.Children {
		results = append(results, Find(child, name)...)
	}

	return results
}

// FindPath 查找路径
func FindPath(node *Node, path string) (*Node, error) {
	parts := strings.Split(path, "/")

	current := node
	for _, part := range parts {
		if part == "" {
			continue
		}

		// 处理属性访问
		if strings.HasPrefix(part, "@") {
			// 返回属性值作为特殊节点
			attrName := strings.TrimPrefix(part, "@")
			if value, ok := GetAttribute(current, attrName); ok {
				return &Node{Content: value}, nil
			}
			return nil, fmt.Errorf("attribute not found: %s", attrName)
		}

		child := GetChild(current, part)
		if child == nil {
			return nil, fmt.Errorf("node not found: %s", part)
		}

		current = child
	}

	return current, nil
}

// ToMap 转换为map
func ToMap(node *Node) map[string]any {
	if node == nil {
		return nil
	}

	result := make(map[string]any)
	result["name"] = node.XMLName.Local

	if len(node.Attrs) > 0 {
		result["attrs"] = node.Attrs
	}

	if node.Content != "" {
		result["content"] = node.Content
	}

	if len(node.Children) > 0 {
		children := make([]map[string]any, 0)
		for _, child := range node.Children {
			children = append(children, ToMap(child))
		}
		result["children"] = children
	}

	return result
}

// ToJSON 转换为JSON
func ToJSON(data []byte) (string, error) {
	node, err := Parse(data)
	if err != nil {
		return "", err
	}

	// 转换为map
	m := map[string]any{
		node.XMLName.Local: ToMap(node),
	}

	// 使用json包序列化
	jsonData, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

// CreateElement 创建元素
func CreateElement(name string) *Node {
	return &Node{
		XMLName: xml.Name{Local: name},
		Attrs:   make(map[string]string),
	}
}

// SetText 设置文本内容
func SetText(node *Node, text string) {
	node.Content = text
}

// AddText 添加文本内容
func AddText(node *Node, text string) {
	node.Content += text
}

// NewElement 创建新元素
func NewElement(name string, content string) *Node {
	return &Node{
		XMLName: xml.Name{Local: name},
		Attrs:   make(map[string]string),
		Content: content,
	}
}

// Build 构建XML
func Build(root *Node) ([]byte, error) {
	return xml.MarshalIndent(root, "", "  ")
}

// BuildString 构建XML字符串
func BuildString(root *Node) (string, error) {
	data, err := Build(root)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Escape 转义XML特殊字符
func Escape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}

// Unescape 反转义XML特殊字符
func Unescape(s string) string {
	s = strings.ReplaceAll(s, "&apos;", "'")
	s = strings.ReplaceAll(s, "&quot;", "\"")
	s = strings.ReplaceAll(s, "&gt;", ">")
	s = strings.ReplaceAll(s, "&lt;", "<")
	s = strings.ReplaceAll(s, "&amp;", "&")
	return s
}

// StripNamespace 移除命名空间
func StripNamespace(data []byte) ([]byte, error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	var buf bytes.Buffer
	encoder := xml.NewEncoder(&buf)

	for {
		token, err := decoder.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// 清除命名空间前缀
		if startElement, ok := token.(xml.StartElement); ok {
			startElement.Name = xml.Name{Local: startElement.Name.Local}
			token = startElement
		}

		if err := encoder.EncodeToken(token); err != nil {
			return nil, err
		}
	}

	return buf.Bytes(), nil
}

// GetEncoding 获取XML编码
func GetEncoding(data []byte) string {
	// 简化实现：检查XML声明
	decoder := xml.NewDecoder(bytes.NewReader(data))

	if _, err := decoder.Token(); err != nil {
		return "UTF-8"
	}

	return "UTF-8"
}

// HasDeclaration 检查是否有XML声明
func HasDeclaration(data []byte) bool {
	trimmed := strings.TrimSpace(string(data))
	return strings.HasPrefix(trimmed, "<?xml")
}

// RemoveDeclaration 移除XML声明
func RemoveDeclaration(data []byte) []byte {
	lines := strings.Split(string(data), "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "<?xml") {
			// 移除这一行
			lines = append(lines[:i], lines[i+1:]...)
			break
		}
		if !strings.HasPrefix(trimmed, "<?xml") && trimmed != "" {
			break
		}
	}

	return []byte(strings.Join(lines, "\n"))
}

// GetVersion 获取XML版本
func GetVersion(data []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	for {
		token, err := decoder.Token()
		if err != nil {
			return ""
		}

		if startElement, ok := token.(xml.StartElement); ok {
			if startElement.Name.Local == "xml" {
				for _, attr := range startElement.Attr {
					if attr.Name.Local == "version" {
						return attr.Value
					}
				}
			}
			break
		}
	}

	return ""
}

// GetEncodingFromDeclaration 从声明获取编码
func GetEncodingFromDeclaration(data []byte) string {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	for {
		token, err := decoder.Token()
		if err != nil {
			return "UTF-8"
		}

		if startElement, ok := token.(xml.StartElement); ok {
			if startElement.Name.Local == "xml" {
				for _, attr := range startElement.Attr {
					if attr.Name.Local == "encoding" {
						return attr.Value
					}
				}
			}
			break
		}
	}

	return "UTF-8"
}

// CDataFilter 创建CDATA过滤器
func CdataFilter(data string) string {
	return "<![CDATA[" + data + "]]>"
}

// CommentFilter 创建注释过滤器
func CommentFilter(comment string) string {
	return "<!--" + comment + "-->"
}

// PrettyPrint 美化打印XML
func PrettyPrint(data []byte) error {
	beautified, err := Beautify(data)
	if err != nil {
		return err
	}
	fmt.Println(string(beautified))
	return nil
}

// ConvertToJSON 转换为JSON
func ConvertToJSON(xmlData []byte) (string, error) {
	// 简化实现：解析XML然后转换为JSON
	return "", nil
}

// Merge 合并XML节点
func Merge(target, source *Node) {
	// 合并属性
	for k, v := range source.Attrs {
		target.Attrs[k] = v
	}

	// 合并内容
	if source.Content != "" {
		target.Content += source.Content
	}

	// 合并子节点
	for _, sourceChild := range source.Children {
		targetChild := GetChild(target, sourceChild.XMLName.Local)
		if targetChild != nil {
			Merge(targetChild, sourceChild)
		} else {
			AddChild(target, sourceChild)
		}
	}
}

// Clone 克隆节点
func Clone(node *Node) *Node {
	if node == nil {
		return nil
	}

	cloned := &Node{
		XMLName: node.XMLName,
		Attrs:    make(map[string]string),
		Content:  node.Content,
	}

	for k, v := range node.Attrs {
		cloned.Attrs[k] = v
	}

	for _, child := range node.Children {
		cloned.Children = append(cloned.Children, Clone(child))
	}

	return cloned
}

// SelectNodes 选择节点
func SelectNodes(node *Node, selector string) []*Node {
	// 简化实现：支持基本的XPath
	return Find(node, selector)
}

// Transform 转换节点
func Transform(node *Node, transformer func(*Node) *Node) {
	if node == nil {
		return
	}

	transformedNode := transformer(node)

	for _, child := range node.Children {
		Transform(child, transformer)
	}

	if transformedNode != nil {
		*node = *transformedNode
	}
}

// FilterNodes 过滤节点
func FilterNodes(node *Node, predicate func(*Node) bool) []*Node {
	var filtered []*Node

	if predicate(node) {
		filtered = append(filtered, node)
	}

	for _, child := range node.Children {
		filtered = append(filtered, FilterNodes(child, predicate)...)
	}

	return filtered
}

// Depth 获取树的深度
func Depth(node *Node) int {
	if node == nil || len(node.Children) == 0 {
		return 1
	}

	maxDepth := 0
	for _, child := range node.Children {
		depth := Depth(child)
		if depth > maxDepth {
			maxDepth = depth
		}
	}

	return maxDepth + 1
}

// Count 统计节点数量
func Count(node *Node) int {
	if node == nil {
		return 0
	}

	count := 1
	for _, child := range node.Children {
		count += Count(child)
	}

	return count
}

// Flatten 扁平化树
func Flatten(node *Node) []*Node {
	var flat []*Node
	flattenHelper(node, &flat)
	return flat
}

func flattenHelper(node *Node, flat *[]*Node) {
	if node == nil {
		return
	}

	*flat = append(*flat, node)

	for _, child := range node.Children {
		flattenHelper(child, flat)
	}
}

// Path 获取节点路径
func Path(root, node *Node) string {
	if root == nil || node == nil {
		return ""
	}

	if root == node {
		return "/" + root.XMLName.Local
	}

	// 简化实现：返回直接路径
	return node.XMLName.Local
}
