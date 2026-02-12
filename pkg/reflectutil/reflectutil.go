// Package reflectutil 提供反射处理工具
package reflectutil

import (
	"fmt"
	"reflect"
	"strings"
)

// GetValue 获取值
func GetValue(obj any, field string) (any, error) {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not a struct")
	}

	field = strings.Title(field)
	f := v.FieldByName(field)

	if !f.IsValid() {
		return nil, fmt.Errorf("field not found: %s", field)
	}

	return f.Interface(), nil
}

// SetValue 设置值
func SetValue(obj any, field string, value any) error {
	v := reflect.ValueOf(obj)

	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("not a pointer to struct")
	}

	v = v.Elem()

	field = strings.Title(field)
	f := v.FieldByName(field)

	if !f.IsValid() {
		return fmt.Errorf("field not found: %s", field)
	}

	if !f.CanSet() {
		return fmt.Errorf("cannot set field: %s", field)
	}

	val := reflect.ValueOf(value)
	if f.Type() != val.Type() {
		return fmt.Errorf("type mismatch: field is %v, value is %v", f.Type(), val.Type())
	}

	f.Set(val)
	return nil
}

// GetFields 获取所有字段名
func GetFields(obj any) []string {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return []string{}
	}

	t := v.Type()
	fields := make([]string, 0, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		// 跳过非导出字段
		if field.PkgPath == "" {
			fields = append(fields, field.Name)
		}
	}

	return fields
}

// GetFieldTags 获取字段标签
func GetFieldTags(obj any, field, tagKey string) (string, error) {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("not a struct")
	}

	t := v.Type()

	field = strings.Title(field)
	f, ok := t.FieldByName(field)
	if !ok {
		return "", fmt.Errorf("field not found: %s", field)
	}

	return f.Tag.Get(tagKey), nil
}

// GetFieldType 获取字段类型
func GetFieldType(obj any, field string) (reflect.Type, error) {
	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not a struct")
	}

	t := v.Type()

	field = strings.Title(field)
	f, ok := t.FieldByName(field)
	if !ok {
		return nil, fmt.Errorf("field not found: %s", field)
	}

	return f.Type, nil
}

// IsNil 检查是否为nil
func IsNil(obj any) bool {
	if obj == nil {
		return true
	}

	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

// IsZero 检查是否为零值
func IsZero(obj any) bool {
	if obj == nil {
		return true
	}

	v := reflect.ValueOf(obj)
	return v.IsZero()
}

// GetKind 获取类型
func GetKind(obj any) reflect.Kind {
	if obj == nil {
		return reflect.Invalid
	}

	v := reflect.ValueOf(obj)
	return v.Kind()
}

// GetTypeName 获取类型名称
func GetTypeName(obj any) string {
	if obj == nil {
		return ""
	}

	t := reflect.TypeOf(obj)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Name()
}

// GetPackageName 获取包名
func GetPackageName(obj any) string {
	if obj == nil {
		return ""
	}

	t := reflect.TypeOf(obj)

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.PkgPath()
}

// HasMethod 检查是否有方法
func HasMethod(obj any, method string) bool {
	if obj == nil {
		return false
	}

	v := reflect.ValueOf(obj)
	m := v.MethodByName(method)

	return m.IsValid()
}

// CallMethod 调用方法
func CallMethod(obj any, method string, args ...any) ([]any, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil object")
	}

	v := reflect.ValueOf(obj)
	m := v.MethodByName(method)

	if !m.IsValid() {
		return nil, fmt.Errorf("method not found: %s", method)
	}

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	out := m.Call(in)

	result := make([]any, len(out))
	for i, v := range out {
		result[i] = v.Interface()
	}

	return result, nil
}

// GetMethods 获取所有方法
func GetMethods(obj any) []string {
	if obj == nil {
		return []string{}
	}

	t := reflect.TypeOf(obj)
	methods := make([]string, 0, t.NumMethod())

	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		methods = append(methods, m.Name)
	}

	return methods
}

// Copy 复制对象
func Copy(obj any) (any, error) {
	if obj == nil {
		return nil, nil
	}

	v := reflect.ValueOf(obj)
	typ := v.Type()

	// 创建新实例
	copy := reflect.New(typ.Elem())

	// 复制字段
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	copyValue := copy.Elem()

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanSet() {
				copyValue.Field(i).Set(field)
			}
		}
	}

	return copy.Interface(), nil
}

// Clone 深度克隆
func Clone(obj any) (any, error) {
	if obj == nil {
		return nil, nil
	}

	v := reflect.ValueOf(obj)

	// 处理指针
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.New(v.Type()).Interface(), nil
		}
		v = v.Elem()
	}

	// 创建新实例
	copy := reflect.New(v.Type())
	copyValue := copy.Elem()

	// 深度复制
	copyValue.Set(deepCopy(v))

	return copy.Interface(), nil
}

func deepCopy(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		copy := reflect.New(v.Type().Elem())
		copy.Elem().Set(deepCopy(v.Elem()))
		return copy

	case reflect.Struct:
		copy := reflect.New(v.Type()).Elem()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			if field.CanSet() {
				copy.Field(i).Set(deepCopy(field))
			}
		}
		return copy

	case reflect.Slice:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		copy := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
		for i := 0; i < v.Len(); i++ {
			copy.Index(i).Set(deepCopy(v.Index(i)))
		}
		return copy

	case reflect.Map:
		if v.IsNil() {
			return reflect.Zero(v.Type())
		}
		copy := reflect.MakeMap(v.Type())
		for _, key := range v.MapKeys() {
			copy.SetMapIndex(deepCopy(key), deepCopy(v.MapIndex(key)))
		}
		return copy

	default:
		return v
	}
}

// ToMap 转换为map
func ToMap(obj any) (map[string]any, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil object")
	}

	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not a struct")
	}

	result := make(map[string]any)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		// 跳过非导出字段
		if field.PkgPath != "" {
			continue
		}

		fieldValue := v.Field(i)
		if fieldValue.CanInterface() {
			result[field.Name] = fieldValue.Interface()
		}
	}

	return result, nil
}

// MapToStruct map转结构体
func MapToStruct(data map[string]any, obj any) error {
	v := reflect.ValueOf(obj)

	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("not a pointer to struct")
	}

	v = v.Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldName := field.Name

		// 检查标签
		if tag := field.Tag.Get("map"); tag != "" {
			if idx := strings.Index(tag, ","); idx != -1 {
				fieldName = tag[:idx]
			} else {
				fieldName = tag
			}
		}

		if value, ok := data[fieldName]; ok {
			fieldValue := v.Field(i)
			if fieldValue.CanSet() {
				val := reflect.ValueOf(value)
				if val.Type().AssignableTo(fieldValue.Type()) {
					fieldValue.Set(val)
				}
			}
		}
	}

	return nil
}

// GetTypeNameByType 获取类型名称
func GetTypeNameByType(t reflect.Type) string {
	if t == nil {
		return ""
	}

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t.Name()
}

// IsPointer 检查是否为指针
func IsPointer(obj any) bool {
	if obj == nil {
		return false
	}

	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Ptr
}

// IsSlice 检查是否为切片
func IsSlice(obj any) bool {
	if obj == nil {
		return false
	}

	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Slice
}

// IsArray 检查是否为数组
func IsArray(obj any) bool {
	if obj == nil {
		return false
	}

	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Array
}

// IsMap 检查是否为map
func IsMap(obj any) bool {
	if obj == nil {
		return false
	}

	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Map
}

// IsStruct 检查是否为结构体
func IsStruct(obj any) bool {
	if obj == nil {
		return false
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	return v.Kind() == reflect.Struct
}

// IsFunc 检查是否为函数
func IsFunc(obj any) bool {
	if obj == nil {
		return false
	}

	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Func
}

// IsInterface 检查是否为接口
func IsInterface(obj any) bool {
	if obj == nil {
		return false
	}

	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Interface
}

// Len 获取长度
func Len(obj any) int {
	if obj == nil {
		return 0
	}

	v := reflect.ValueOf(obj)
	switch v.Kind() {
	case reflect.String, reflect.Array, reflect.Slice, reflect.Map, reflect.Chan:
		return v.Len()
	default:
		return 0
	}
}

// Iterate 遍历
func Iterate(obj any) ([]any, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil object")
	}

	v := reflect.ValueOf(obj)

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		result := make([]any, v.Len())
		for i := 0; i < v.Len(); i++ {
			result[i] = v.Index(i).Interface()
		}
		return result, nil

	case reflect.Map:
		result := make([]any, 0, v.Len())
		for _, key := range v.MapKeys() {
			result = append(result, key.Interface())
		}
		return result, nil

	default:
		return nil, fmt.Errorf("not iterable")
	}
}

// GetSliceLength 获取切片长度
func GetSliceLength(obj any) (int, error) {
	if obj == nil {
		return 0, fmt.Errorf("nil object")
	}

	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return 0, fmt.Errorf("not a slice or array")
	}

	return v.Len(), nil
}

// GetSliceElement 获取切片元素
func GetSliceElement(obj any, index int) (any, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil object")
	}

	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		return nil, fmt.Errorf("not a slice or array")
	}

	if index < 0 || index >= v.Len() {
		return nil, fmt.Errorf("index out of bounds: %d", index)
	}

	return v.Index(index).Interface(), nil
}

// ConvertType 转换类型
func ConvertType(value any, targetType reflect.Type) (any, error) {
	if value == nil {
		return reflect.Zero(targetType).Interface(), nil
	}

	v := reflect.ValueOf(value)

	if v.Type().ConvertibleTo(targetType) {
		return v.Convert(targetType).Interface(), nil
	}

	return nil, fmt.Errorf("cannot convert %v to %v", v.Type(), targetType)
}

// Implements 检查是否实现接口
func Implements(obj any, iface any) bool {
	if obj == nil || iface == nil {
		return false
	}

	objType := reflect.TypeOf(obj)
	ifaceType := reflect.TypeOf(iface)

	if ifaceType.Kind() != reflect.Interface {
		return false
	}

	return objType.Implements(ifaceType)
}

// Compare 比较两个值
func Compare(a, b any) int {
	if a == nil && b == nil {
		return 0
	}
	if a == nil {
		return -1
	}
	if b == nil {
		return 1
	}

	va := reflect.ValueOf(a)
	vb := reflect.ValueOf(b)

	// 简化实现，实际应该处理更多类型
	switch va.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		aInt := va.Int()
		bInt := vb.Int()
		if aInt < bInt {
			return -1
		} else if aInt > bInt {
			return 1
		}
		return 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		aUint := va.Uint()
		bUint := vb.Uint()
		if aUint < bUint {
			return -1
		} else if aUint > bUint {
			return 1
		}
		return 0

	case reflect.Float32, reflect.Float64:
		aFloat := va.Float()
		bFloat := vb.Float()
		if aFloat < bFloat {
			return -1
		} else if aFloat > bFloat {
			return 1
		}
		return 0

	case reflect.String:
		aStr := va.String()
		bStr := vb.String()
		if aStr < bStr {
			return -1
		} else if aStr > bStr {
			return 1
		}
		return 0

	default:
		return 0
	}
}

// Equal 检查是否相等
func Equal(a, b any) bool {
	return Compare(a, b) == 0
}

// GetElem 获取元素
func GetElem(obj any) any {
	if obj == nil {
		return nil
	}

	v := reflect.ValueOf(obj)

	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	return v.Interface()
}

// MakeSlice 创建切片
func MakeSlice(elemType reflect.Type, length, capacity int) (any, error) {
	slice := reflect.MakeSlice(reflect.SliceOf(elemType), length, capacity)
	return slice.Interface(), nil
}

// MakeMap 创建map
func MakeMap(keyType, valueType reflect.Type) (any, error) {
	m := reflect.MakeMap(reflect.MapOf(keyType, valueType))
	return m.Interface(), nil
}

// New 创建新实例
func New(typ reflect.Type) (any, error) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	return reflect.New(typ).Interface(), nil
}

// CallFunc 调用函数
func CallFunc(fn any, args ...any) ([]any, error) {
	if fn == nil {
		return nil, fmt.Errorf("nil function")
	}

	v := reflect.ValueOf(fn)
	if v.Kind() != reflect.Func {
		return nil, fmt.Errorf("not a function")
	}

	in := make([]reflect.Value, len(args))
	for i, arg := range args {
		in[i] = reflect.ValueOf(arg)
	}

	out := v.Call(in)

	result := make([]any, len(out))
	for i, v := range out {
		result[i] = v.Interface()
	}

	return result, nil
}

// GetStructTags 获取结构体标签
func GetStructTags(obj any, tagKey string) (map[string]string, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil object")
	}

	v := reflect.ValueOf(obj)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("not a struct")
	}

	t := v.Type()
	result := make(map[string]string)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get(tagKey)
		if tag != "" {
			result[field.Name] = tag
		}
	}

	return result, nil
}

// FillStruct 用map填充结构体
func FillStruct(obj any, data map[string]any) error {
	return MapToStruct(data, obj)
}

// GetFieldValue 获取字段值（支持嵌套）
func GetFieldValue(obj any, fieldPath string) (any, error) {
	if obj == nil {
		return nil, fmt.Errorf("nil object")
	}

	parts := strings.Split(fieldPath, ".")
	v := reflect.ValueOf(obj)

	for _, part := range parts {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return nil, fmt.Errorf("not a struct at: %s", part)
		}

		part = strings.Title(part)
		field := v.FieldByName(part)

		if !field.IsValid() {
			return nil, fmt.Errorf("field not found: %s", part)
		}

		v = field
	}

	return v.Interface(), nil
}

// SetFieldValue 设置字段值（支持嵌套）
func SetFieldValue(obj any, fieldPath string, value any) error {
	if obj == nil {
		return fmt.Errorf("nil object")
	}

	parts := strings.Split(fieldPath, ".")
	v := reflect.ValueOf(obj)

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("not a pointer to struct")
	}

	// 最后一部分用于设置值
	lastPart := parts[len(parts)-1]
	parts = parts[:len(parts)-1]

	// 导航到父结构
	for _, part := range parts {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return fmt.Errorf("not a struct at: %s", part)
		}

		part = strings.Title(part)
		field := v.FieldByName(part)

		if !field.IsValid() {
			return fmt.Errorf("field not found: %s", part)
		}

		v = field
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct")
	}

	lastPart = strings.Title(lastPart)
	field := v.FieldByName(lastPart)

	if !field.IsValid() {
		return fmt.Errorf("field not found: %s", lastPart)
	}

	if !field.CanSet() {
		return fmt.Errorf("cannot set field: %s", lastPart)
	}

	val := reflect.ValueOf(value)
	if field.Type() != val.Type() {
		return fmt.Errorf("type mismatch: field is %v, value is %v", field.Type(), val.Type())
	}

	field.Set(val)
	return nil
}
