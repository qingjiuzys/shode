// Package i18n 提供国际化功能。
package i18n

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Locale 语言环境
type Locale struct {
	Code   string
	Name   string
	Region string
}

// CommonLocales 常用语言环境
var CommonLocales = map[string]Locale{
	"en": {Code: "en", Name: "English", Region: "US"},
	"zh": {Code: "zh", Name: "中文", Region: "CN"},
	"ja": {Code: "ja", Name: "日本語", Region: "JP"},
	"ko": {Code: "ko", Name: "한국어", Region: "KR"},
	"es": {Code: "es", Name: "Español", Region: "ES"},
	"fr": {Code: "fr", Name: "Français", Region: "FR"},
	"de": {Code: "de", Name: "Deutsch", Region: "DE"},
	"ru": {Code: "ru", Name: "Русский", Region: "RU"},
}

// Translator 翻译器接口
type Translator interface {
	Translate(ctx context.Context, key string, args ...interface{}) string
	Locale(ctx context.Context) Locale
}

// I18n 国际化
type I18n struct {
	defaultLocale  Locale
	currentLocale  Locale
	translations   map[string]map[string]string // locale -> key -> translation
	formatters     map[string]*Formatter       // locale -> formatter
	mu             sync.RWMutex
	fallbackChain  []string
}

// Formatter 格式化器
type Formatter struct {
	DateShort     string
	DateLong      string
	TimeShort     string
	TimeLong      string
	DateTimeShort string
	DateTimeLong  string
	NumberFormat  string
	CurrencyFormat string
}

// NewI18n 创建国际化实例
func NewI18n(defaultLocale string) *I18n {
	locale, exists := CommonLocales[defaultLocale]
	if !exists {
		locale = CommonLocales["en"]
	}

	return &I18n{
		defaultLocale: locale,
		currentLocale: locale,
		translations:  make(map[string]map[string]string),
		formatters:    make(map[string]*Formatter),
		fallbackChain: []string{defaultLocale, "en"}, // 回退链
	}
}

// SetLocale 设置当前语言环境
func (i *I18n) SetLocale(localeCode string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	locale, exists := CommonLocales[localeCode]
	if !exists {
		return fmt.Errorf("unsupported locale: %s", localeCode)
	}

	i.currentLocale = locale

	// 初始化格式化器（如果不存在）
	if _, exists := i.formatters[localeCode]; !exists {
		i.formatters[localeCode] = i.getDefaultFormatter(localeCode)
	}

	return nil
}

// GetLocale 获取当前语言环境
func (i *I18n) GetLocale() Locale {
	i.mu.RLock()
	defer i.mu.RUnlock()

	return i.currentLocale
}

// Translate 翻译文本
func (i *I18n) Translate(ctx context.Context, key string, args ...interface{}) string {
	i.mu.RLock()
	defer i.mu.RUnlock()

	localeCode := i.currentLocale.Code

	// 查找翻译
	if translations, exists := i.translations[localeCode]; exists {
		if translation, exists := translations[key]; exists {
			if len(args) > 0 {
				return fmt.Sprintf(translation, args...)
			}
			return translation
		}
	}

	// 尝试回退链
	for _, fallbackCode := range i.fallbackChain {
		if translations, exists := i.translations[fallbackCode]; exists {
			if translation, exists := translations[key]; exists {
				if len(args) > 0 {
					return fmt.Sprintf(translation, args...)
				}
				return translation
			}
		}
	}

	// 返回原键
	if len(args) > 0 {
		return fmt.Sprintf(key, args...)
	}
	return key
}

// AddTranslation 添加翻译
func (i *I18n) AddTranslation(localeCode, key, translation string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, exists := i.translations[localeCode]; !exists {
		i.translations[localeCode] = make(map[string]string)
	}

	i.translations[localeCode][key] = translation
}

// AddTranslations 批量添加翻译
func (i *I18n) AddTranslations(localeCode string, translations map[string]string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if _, exists := i.translations[localeCode]; !exists {
		i.translations[localeCode] = make(map[string]string)
	}

	for key, translation := range translations {
		i.translations[localeCode][key] = translation
	}
}

// LoadTranslationsFromFile 从文件加载翻译
func (i *I18n) LoadTranslationsFromFile(localeCode, filePath string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read translation file: %w", err)
	}

	var translations map[string]string
	if err := json.Unmarshal(data, &translations); err != nil {
		return fmt.Errorf("failed to parse translation file: %w", err)
	}

	if _, exists := i.translations[localeCode]; !exists {
		i.translations[localeCode] = make(map[string]string)
	}

	for key, translation := range translations {
		i.translations[localeCode][key] = translation
	}

	return nil
}

// LoadTranslationsFromDir 从目录加载翻译
func (i *I18n) LoadTranslationsFromDir(dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		return fmt.Errorf("failed to read translation files: %w", err)
	}

	for _, file := range files {
		// 文件名格式: en.json, zh.json
		filename := filepath.Base(file)
		localeCode := filename[:len(filename)-5]

		if err := i.LoadTranslationsFromFile(localeCode, file); err != nil {
			return fmt.Errorf("failed to load translations for %s: %w", localeCode, err)
		}
	}

	return nil
}

// FormatDate 格式化日期
func (i *I18n) FormatDate(date time.Time, format string) string {
	i.mu.RLock()
	formatter, exists := i.formatters[i.currentLocale.Code]
	i.mu.RUnlock()

	if !exists {
		formatter = i.getDefaultFormatter(i.currentLocale.Code)
	}

	layout := formatter.DateLong
	if format == "short" {
		layout = formatter.DateShort
	}

	return date.Format(layout)
}

// FormatTime 格式化时间
func (i *I18n) FormatTime(t time.Time, format string) string {
	i.mu.RLock()
	formatter, exists := i.formatters[i.currentLocale.Code]
	i.mu.RUnlock()

	if !exists {
		formatter = i.getDefaultFormatter(i.currentLocale.Code)
	}

	layout := formatter.TimeLong
	if format == "short" {
		layout = formatter.TimeShort
	}

	return t.Format(layout)
}

// FormatDateTime 格式化日期时间
func (i *I18n) FormatDateTime(dt time.Time, format string) string {
	i.mu.RLock()
	formatter, exists := i.formatters[i.currentLocale.Code]
	i.mu.RUnlock()

	if !exists {
		formatter = i.getDefaultFormatter(i.currentLocale.Code)
	}

	layout := formatter.DateTimeLong
	if format == "short" {
		layout = formatter.DateTimeShort
	}

	return dt.Format(layout)
}

// FormatNumber 格式化数字
func (i *I18n) FormatNumber(n float64) string {
	// 简化实现，实际应该根据语言环境格式化
	return fmt.Sprintf("%.2f", n)
}

// FormatCurrency 格式化货币
func (i *I18n) FormatCurrency(amount float64, currency string) string {
	// 简化实现，实际应该根据语言环境和货币符号格式化
	return fmt.Sprintf("%s %.2f", currency, amount)
}

// getDefaultFormatter 获取默认格式化器
func (i *I18n) getDefaultFormatter(localeCode string) *Formatter {
	// 根据语言环境返回不同的格式
	switch localeCode {
	case "zh":
		return &Formatter{
			DateShort:     "2006-01-02",
			DateLong:      "2006年01月02日",
			TimeShort:     "15:04",
			TimeLong:      "15:04:05",
			DateTimeShort: "2006-01-02 15:04",
			DateTimeLong:  "2006年01月02日 15:04:05",
			NumberFormat:  "#,##0.##",
			CurrencyFormat: "¥#,##0.00",
		}
	case "ja":
		return &Formatter{
			DateShort:     "2006/01/02",
			DateLong:      "2006年01月02日",
			TimeShort:     "15:04",
			TimeLong:      "15:04:05",
			DateTimeShort: "2006/01/02 15:04",
			DateTimeLong:  "2006年01月02日 15:04:05",
			NumberFormat:  "#,##0.##",
			CurrencyFormat: "¥#,##0.00",
		}
	case "en":
		fallthrough
	default:
		return &Formatter{
			DateShort:     "01/02/2006",
			DateLong:      "January 2, 2006",
			TimeShort:     "3:04 PM",
			TimeLong:      "3:04:05 PM",
			DateTimeShort: "01/02/2006 3:04 PM",
			DateTimeLong:  "January 2, 2006 3:04:05 PM",
			NumberFormat:  "#,##0.##",
			CurrencyFormat: "$#,##0.00",
		}
	}
}

// SetFallbackChain 设置回退链
func (i *I18n) SetFallbackChain(chain []string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.fallbackChain = chain
}

// GetSupportedLocales 获取支持的语言环境
func (i *I18n) GetSupportedLocales() []Locale {
	i.mu.RLock()
	defer i.mu.RUnlock()

	locales := make([]Locale, 0, len(i.translations))
	for code := range i.translations {
		if locale, exists := CommonLocales[code]; exists {
			locales = append(locales, locale)
		}
	}

	return locales
}

// ContextKey 上下文键类型
type ContextKey int

const (
	LocaleKey ContextKey = iota
)

// NewContextWithLocale 创建带语言环境的上下文
func NewContextWithLocale(ctx context.Context, locale string) context.Context {
	return context.WithValue(ctx, LocaleKey, locale)
}

// LocaleFromContext 从上下文获取语言环境
func LocaleFromContext(ctx context.Context) string {
	if locale, ok := ctx.Value(LocaleKey).(string); ok {
		return locale
	}
	return "en" // 默认英语
}

// GlobalI18n 全局国际化实例
var GlobalI18n = NewI18n("en")

// T 翻译辅助函数（使用全局实例）
func T(key string, args ...interface{}) string {
	return GlobalI18n.Translate(context.Background(), key, args...)
}

// Tf 翻译辅助函数（带格式化）
func Tf(key string, args ...interface{}) string {
	return GlobalI18n.Translate(context.Background(), key, args...)
}

// SetL 设置语言环境（使用全局实例）
func SetL(localeCode string) error {
	return GlobalI18n.SetLocale(localeCode)
}

// GetL 获取当前语言环境（使用全局实例）
func GetL() Locale {
	return GlobalI18n.GetLocale()
}

// LoadTranslations 加载翻译（使用全局实例）
func LoadTranslations(localeCode string, translations map[string]string) {
	GlobalI18n.AddTranslations(localeCode, translations)
}

// Timezone 时区处理
type Timezone struct {
	Location *time.Location
	Name     string
	Offset   int // 偏移量（秒）
}

// CommonTimezones 常用时区
var CommonTimezones = map[string]Timezone{
	"UTC": {
		Location: time.UTC,
		Name:     "UTC",
		Offset:   0,
	},
	"America/New_York": {
		Location: mustLoadLocation("America/New_York"),
		Name:     "EST/EDT",
		Offset:   -5 * 60 * 60,
	},
	"America/Los_Angeles": {
		Location: mustLoadLocation("America/Los_Angeles"),
		Name:     "PST/PDT",
		Offset:   -8 * 60 * 60,
	},
	"Europe/London": {
		Location: mustLoadLocation("Europe/London"),
		Name:     "GMT/BST",
		Offset:   0,
	},
	"Asia/Shanghai": {
		Location: mustLoadLocation("Asia/Shanghai"),
		Name:     "CST",
		Offset:   8 * 60 * 60,
	},
	"Asia/Tokyo": {
		Location: mustLoadLocation("Asia/Tokyo"),
		Name:     "JST",
		Offset:   9 * 60 * 60,
	},
}

// mustLoadLocation 加载时区位置
func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return time.UTC
	}
	return loc
}

// ConvertTimezone 转换时区
func ConvertTimezone(t time.Time, from, to string) (time.Time, error) {
	_, exists := CommonTimezones[from]
	if !exists {
		return time.Time{}, fmt.Errorf("unknown timezone: %s", from)
	}

	toTz, exists := CommonTimezones[to]
	if !exists {
		return time.Time{}, fmt.Errorf("unknown timezone: %s", to)
	}

	// 转换到目标时区
	return t.In(toTz.Location), nil
}

// FormatInTimezone 在指定时区格式化时间
func FormatInTimezone(t time.Time, timezone, format string) (string, error) {
	tz, exists := CommonTimezones[timezone]
	if !exists {
		return "", fmt.Errorf("unknown timezone: %s", timezone)
	}

	localTime := t.In(tz.Location)
	return localTime.Format(format), nil
}
