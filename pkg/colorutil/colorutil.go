// Package colorutil 提供颜色处理工具
package colorutil

import (
	"fmt"
	"math"
	mrand "math/rand"
	"strconv"
	"strings"
	"time"
)

// RGB RGB颜色
type RGB struct {
	R, G, B uint8
}

// RGBA RGBA颜色
type RGBA struct {
	R, G, B, A uint8
}

// HSL HSL颜色
type HSL struct {
	H, S, L float64
}

// HSV HSV颜色
type HSV struct {
	H, S, V float64
}

// CMYK CMYK颜色
type CMYK struct {
	C, M, Y, K float64
}

// NewRGB 创建RGB颜色
func NewRGB(r, g, b uint8) RGB {
	return RGB{R: r, G: g, B: b}
}

// NewRGBA 创建RGBA颜色
func NewRGBA(r, g, b, a uint8) RGBA {
	return RGBA{R: r, G: g, B: b, A: a}
}

// NewHSL 创建HSL颜色
func NewHSL(h, s, l float64) HSL {
	return HSL{H: h, S: s, L: l}
}

// NewHSV 创建HSV颜色
func NewHSV(h, s, v float64) HSV {
	return HSV{H: h, S: s, V: v}
}

// HexToRGB 十六进制转RGB
func HexToRGB(hex string) (RGB, error) {
	hex = strings.TrimPrefix(hex, "#")

	if len(hex) != 6 {
		return RGB{}, fmt.Errorf("invalid hex color: %s", hex)
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return RGB{}, err
	}

	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return RGB{}, err
	}

	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return RGB{}, err
	}

	return RGB{R: uint8(r), G: uint8(g), B: uint8(b)}, nil
}

// RGBToHex RGB转十六进制
func RGBToHex(rgb RGB) string {
	return fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
}

// RGBToHSL RGB转HSL
func RGBToHSL(rgb RGB) HSL {
	r := float64(rgb.R) / 255
	g := float64(rgb.G) / 255
	b := float64(rgb.B) / 255

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	var h, s, l float64

	l = (max + min) / 2

	if delta == 0 {
		h = 0
		s = 0
	} else {
		s = delta / (1 - math.Abs(2*l-1))

		if max == r {
			h = math.Mod((g-b)/delta, 6)
		} else if max == g {
			h = (b-r)/delta + 2
		} else {
			h = (r-g)/delta + 4
		}

		h *= 60
		if h < 0 {
			h += 360
		}
	}

	return HSL{H: h, S: s, L: l}
}

// HSLToRGB HSL转RGB
func HSLToRGB(hsl HSL) RGB {
	h := hsl.H
	s := hsl.S
	l := hsl.L

	if s == 0 {
		gray := uint8(l * 255)
		return RGB{R: gray, G: gray, B: gray}
	}

	var r, g, b float64

	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2

	switch {
	case 0 <= h && h < 60:
		r, g, b = c, x, 0
	case 60 <= h && h < 120:
		r, g, b = x, c, 0
	case 120 <= h && h < 180:
		r, g, b = 0, c, x
	case 180 <= h && h < 240:
		r, g, b = 0, x, c
	case 240 <= h && h < 300:
		r, g, b = x, 0, c
	case 300 <= h && h < 360:
		r, g, b = c, 0, x
	}

	return RGB{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
	}
}

// RGBToHSV RGB转HSV
func RGBToHSV(rgb RGB) HSV {
	r := float64(rgb.R) / 255
	g := float64(rgb.G) / 255
	b := float64(rgb.B) / 255

	max := math.Max(r, math.Max(g, b))
	min := math.Min(r, math.Min(g, b))
	delta := max - min

	var h, s, v float64

	v = max

	if delta == 0 {
		h = 0
		s = 0
	} else {
		s = delta / max

		if max == r {
			h = math.Mod((g-b)/delta, 6)
		} else if max == g {
			h = (b-r)/delta + 2
		} else {
			h = (r-g)/delta + 4
		}

		h *= 60
		if h < 0 {
			h += 360
		}
	}

	return HSV{H: h, S: s, V: v}
}

// HSVToRGB HSV转RGB
func HSVToRGB(hsv HSV) RGB {
	h := hsv.H
	s := hsv.S
	v := hsv.V

	c := v * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := v - c

	var r, g, b float64

	switch {
	case 0 <= h && h < 60:
		r, g, b = c, x, 0
	case 60 <= h && h < 120:
		r, g, b = x, c, 0
	case 120 <= h && h < 180:
		r, g, b = 0, c, x
	case 180 <= h && h < 240:
		r, g, b = 0, x, c
	case 240 <= h && h < 300:
		r, g, b = x, 0, c
	case 300 <= h && h < 360:
		r, g, b = c, 0, x
	}

	return RGB{
		R: uint8((r + m) * 255),
		G: uint8((g + m) * 255),
		B: uint8((b + m) * 255),
	}
}

// RGBToCMYK RGB转CMYK
func RGBToCMYK(rgb RGB) CMYK {
	r := float64(rgb.R) / 255
	g := float64(rgb.G) / 255
	b := float64(rgb.B) / 255

	k := 1 - math.Max(r, math.Max(g, b))
	c := (1 - r - k) / (1 - k)
	m := (1 - g - k) / (1 - k)
	y := (1 - b - k) / (1 - k)

	if k == 1 {
		c, m, y = 0, 0, 0
	}

	return CMYK{C: c, M: m, Y: y, K: k}
}

// CMYKToRGB CMYK转RGB
func CMYKToRGB(cmyk CMYK) RGB {
	c := cmyk.C
	m := cmyk.M
	y := cmyk.Y
	k := cmyk.K

	r := (1 - c) * (1 - k)
	g := (1 - m) * (1 - k)
	b := (1 - y) * (1 - k)

	return RGB{
		R: uint8(r * 255),
		G: uint8(g * 255),
		B: uint8(b * 255),
	}
}

// ToRGBA 转为RGBA
func (rgb RGB) ToRGBA(a uint8) RGBA {
	return RGBA{R: rgb.R, G: rgb.G, B: rgb.B, A: a}
}

// String RGB转字符串
func (rgb RGB) String() string {
	return RGBToHex(rgb)
}

// String RGBA转字符串
func (rgba RGBA) String() string {
	return fmt.Sprintf("rgba(%d,%d,%d,%f)", rgba.R, rgba.G, rgba.B, float64(rgba.A)/255)
}

// Blend 混合两种颜色
func Blend(c1, c2 RGB, ratio float64) RGB {
	r := uint8(float64(c1.R)*(1-ratio) + float64(c2.R)*ratio)
	g := uint8(float64(c1.G)*(1-ratio) + float64(c2.G)*ratio)
	b := uint8(float64(c1.B)*(1-ratio) + float64(c2.B)*ratio)
	return RGB{R: r, G: g, B: b}
}

// Invert 反转颜色
func Invert(rgb RGB) RGB {
	return RGB{
		R: 255 - rgb.R,
		G: 255 - rgb.G,
		B: 255 - rgb.B,
	}
}

// Grayscale 灰度化
func Grayscale(rgb RGB) RGB {
	gray := uint8(0.299*float64(rgb.R) + 0.587*float64(rgb.G) + 0.114*float64(rgb.B))
	return RGB{R: gray, G: gray, B: gray}
}

// Brightness 调整亮度
func Brightness(rgb RGB, factor float64) RGB {
	r := uint8(math.Min(255, math.Max(0, float64(rgb.R)*factor)))
	g := uint8(math.Min(255, math.Max(0, float64(rgb.G)*factor)))
	b := uint8(math.Min(255, math.Max(0, float64(rgb.B)*factor)))
	return RGB{R: r, G: g, B: b}
}

// Lighten 提亮颜色
func Lighten(rgb RGB, amount float64) RGB {
	hsl := RGBToHSL(rgb)
	hsl.L = math.Min(1, hsl.L+amount)
	return HSLToRGB(hsl)
}

// Darken 变暗颜色
func Darken(rgb RGB, amount float64) RGB {
	hsl := RGBToHSL(rgb)
	hsl.L = math.Max(0, hsl.L-amount)
	return HSLToRGB(hsl)
}

// Saturate 调整饱和度
func Saturate(rgb RGB, amount float64) RGB {
	hsl := RGBToHSL(rgb)
	hsl.S = math.Min(1, hsl.S+amount)
	return HSLToRGB(hsl)
}

// Desaturate 降低饱和度
func Desaturate(rgb RGB, amount float64) RGB {
	hsl := RGBToHSL(rgb)
	hsl.S = math.Max(0, hsl.S-amount)
	return HSLToRGB(hsl)
}

// AdjustHue 调整色相
func AdjustHue(rgb RGB, degrees float64) RGB {
	hsl := RGBToHSL(rgb)
	hsl.H = math.Mod(hsl.H+degrees, 360)
	return HSLToRGB(hsl)
}

// Complementary 互补色
func Complementary(rgb RGB) RGB {
	return AdjustHue(rgb, 180)
}

// Triadic 三色组
func Triadic(rgb RGB) [3]RGB {
	return [3]RGB{
		rgb,
		AdjustHue(rgb, 120),
		AdjustHue(rgb, 240),
	}
}

// Analogous 类似色
func Analogous(rgb RGB) [3]RGB {
	return [3]RGB{
		AdjustHue(rgb, -30),
		rgb,
		AdjustHue(rgb, 30),
	}
}

// SplitComplementary 分裂互补色
func SplitComplementary(rgb RGB) [3]RGB {
	comp := AdjustHue(rgb, 180)
	return [3]RGB{
		rgb,
		AdjustHue(comp, -30),
		AdjustHue(comp, 30),
	}
}

// Luminance 亮度
func Luminance(rgb RGB) float64 {
	return 0.299*float64(rgb.R) + 0.587*float64(rgb.G) + 0.114*float64(rgb.B)
}

// ContrastRatio 对比度
func ContrastRatio(c1, c2 RGB) float64 {
	l1 := Luminance(c1)
	l2 := Luminance(c2)

	lighter := math.Max(l1, l2)
	darker := math.Min(l1, l2)

	return (lighter + 0.05) / (darker + 0.05)
}

// IsLight 是否为浅色
func IsLight(rgb RGB) bool {
	return Luminance(rgb) > 128
}

// IsDark 是否为深色
func IsDark(rgb RGB) bool {
	return Luminance(rgb) <= 128
}

// ParseColor 解析颜色字符串
func ParseColor(color string) (RGB, error) {
	// 十六进制
	if strings.HasPrefix(color, "#") {
		return HexToRGB(color)
	}

	// rgb()
	if strings.HasPrefix(color, "rgb(") {
		parts := strings.Split(strings.TrimSuffix(strings.TrimPrefix(color, "rgb("), ")"), ",")
		if len(parts) != 3 {
			return RGB{}, fmt.Errorf("invalid rgb color: %s", color)
		}

		r, _ := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 8)
		g, _ := strconv.ParseUint(strings.TrimSpace(parts[1]), 10, 8)
		b, _ := strconv.ParseUint(strings.TrimSpace(parts[2]), 10, 8)

		return RGB{R: uint8(r), G: uint8(g), B: uint8(b)}, nil
	}

	// 颜色名称
	return colorFromName(color)
}

// colorFromName 从颜色名称获取RGB
func colorFromName(name string) (RGB, error) {
	switch strings.ToLower(name) {
	case "black", "#000":
		return RGB{R: 0, G: 0, B: 0}, nil
	case "white", "#fff":
		return RGB{R: 255, G: 255, B: 255}, nil
	case "red":
		return RGB{R: 255, G: 0, B: 0}, nil
	case "green":
		return RGB{R: 0, G: 255, B: 0}, nil
	case "blue":
		return RGB{R: 0, G: 0, B: 255}, nil
	case "yellow":
		return RGB{R: 255, G: 255, B: 0}, nil
	case "cyan":
		return RGB{R: 0, G: 255, B: 255}, nil
	case "magenta":
		return RGB{R: 255, G: 0, B: 255}, nil
	case "gray", "grey":
		return RGB{R: 128, G: 128, B: 128}, nil
	case "orange":
		return RGB{R: 255, G: 165, B: 0}, nil
	case "purple":
		return RGB{R: 128, G: 0, B: 128}, nil
	case "pink":
		return RGB{R: 255, G: 192, B: 203}, nil
	default:
		return RGB{}, fmt.Errorf("unknown color name: %s", name)
	}
}


// RandomRGB 随机RGB颜色
func RandomRGB() RGB {
	return RGB{
		R: uint8(mrand.Intn(256)),
		G: uint8(mrand.Intn(256)),
		B: uint8(mrand.Intn(256)),
	}
}

// RandomHSL 随机HSL颜色
func RandomHSL() HSL {
	return HSL{
		H: mrand.Float64() * 360,
		S: mrand.Float64(),
		L: mrand.Float64(),
	}
}

// WarmColor 暖色调
func WarmColor() RGB {
	h := mrand.Float64() * 60
	s := 0.5 + mrand.Float64()*0.5
	l := 0.3 + mrand.Float64()*0.4
	hsl := HSL{H: h, S: s, L: l}
	return HSLToRGB(hsl)
}

// CoolColor 冷色调
func CoolColor() RGB {
	h := 180 + mrand.Float64()*120
	s := 0.5 + mrand.Float64()*0.5
	l := 0.3 + mrand.Float64()*0.4
	hsl := HSL{H: h, S: s, L: l}
	return HSLToRGB(hsl)
}

// PastelColor 柔和色调
func PastelColor() RGB {
	h := mrand.Float64() * 360
	hsl := HSL{H: h, S: 0.3, L: 0.8}
	return HSLToRGB(hsl)
}

// EarthColor 大地色调
func EarthColor() RGB {
	h := 30 + mrand.Float64()*30
	s := 0.3 + mrand.Float64()*0.3
	l := 0.2 + mrand.Float64()*0.3
	hsl := HSL{H: h, S: s, L: l}
	return HSLToRGB(hsl)
}

// NeonColor 霓虹色调
func NeonColor() RGB {
	h := mrand.Float64() * 360
	hsl := HSL{H: h, S: 1, L: 0.5}
	return HSLToRGB(hsl)
}

func init() {
	mrand.Seed(time.Now().UnixNano())
}
