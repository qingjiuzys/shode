package cookie

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CookieManager Cookie 管理器
type CookieManager struct{}

// NewCookieManager 创建 Cookie 管理器
func NewCookieManager() *CookieManager {
	return &CookieManager{}
}

// SetCookie 设置 Cookie
func (cm *CookieManager) SetCookie(w http.ResponseWriter, name, value string, options string) error {
	cookie := &http.Cookie{
		Name:  name,
		Value: value,
	}
	
	// 解析选项
	if options != "" {
		parts := strings.Split(options, ";")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			
			if strings.HasPrefix(part, "Path=") {
				cookie.Path = strings.TrimPrefix(part, "Path=")
			} else if strings.HasPrefix(part, "Domain=") {
				cookie.Domain = strings.TrimPrefix(part, "Domain=")
			} else if strings.HasPrefix(part, "Max-Age=") {
				maxAgeStr := strings.TrimPrefix(part, "Max-Age=")
				maxAge, _ := strconv.Atoi(maxAgeStr)
				cookie.MaxAge = maxAge
			} else if part == "Secure" {
				cookie.Secure = true
			} else if part == "HttpOnly" {
				cookie.HttpOnly = true
			}
		}
	}
	
	http.SetCookie(w, cookie)
	return nil
}

// GetCookie 获取 Cookie
func (cm *CookieManager) GetCookie(r *http.Request, name string) (string, error) {
	cookie, err := r.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// DeleteCookie 删除 Cookie
func (cm *CookieManager) DeleteCookie(w http.ResponseWriter, name string, path string) error {
	cookie := &http.Cookie{
		Name:   name,
		Value:  "",
		Path:   path,
		MaxAge: -1,
	}
	
	http.SetCookie(w, cookie)
	return nil
}
