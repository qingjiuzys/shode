// Package netutil 提供网络工具函数
package netutil

import (
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// GetLocalIP 获取本地IP地址
func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

// GetPublicIP 获取公网IP地址
func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get public IP: status %d", resp.StatusCode)
	}

	ip := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			ip = append(ip, buf[:n]...)
		}
		if err != nil {
			break
		}
	}

	return string(ip), nil
}

// IsValidIP 检查是否为有效IP地址
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidIPv4 检查是否为有效IPv4地址
func IsValidIPv4(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.To4() != nil
}

// IsValidIPv6 检查是否为有效IPv6地址
func IsValidIPv6(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	return parsedIP.To4() == nil
}

// IPToLong IP地址转长整型
func IPToLong(ip string) (uint32, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ip)
	}

	ipv4 := parsedIP.To4()
	if ipv4 == nil {
		return 0, fmt.Errorf("not an IPv4 address: %s", ip)
	}

	return binary.BigEndian.Uint32(ipv4), nil
}

// LongToIP 长整型转IP地址
func LongToIP(long uint32) string {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, long)
	return ip.String()
}

// IPToInt64 IP地址转int64
func IPToInt64(ip string) (int64, error) {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return 0, fmt.Errorf("invalid IP address: %s", ip)
	}

	ipv4 := parsedIP.To4()
	if ipv4 == nil {
		return 0, fmt.Errorf("not an IPv4 address: %s", ip)
	}

	return int64(binary.BigEndian.Uint32(ipv4)), nil
}

// Int64ToIP int64转IP地址
func Int64ToIP(long int64) string {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, uint32(long))
	return ip.String()
}

// IPRange 获取IP范围
func IPRange(startIP, endIP string) ([]string, error) {
	start := net.ParseIP(startIP)
	end := net.ParseIP(endIP)

	if start == nil || end == nil {
		return nil, fmt.Errorf("invalid IP addresses")
	}

	start = start.To4()
	end = end.To4()

	if start == nil || end == nil {
		return nil, fmt.Errorf("only IPv4 addresses supported")
	}

	var ips []string
	for ip := start; !ip.Equal(end); {
		ips = append(ips, ip.String())
		ip = incIP(ip)
	}
	ips = append(ips, end.String())

	return ips, nil
}

func incIP(ip net.IP) net.IP {
	ip = ip.To4()
	if ip == nil {
		return ip
	}

	for i := len(ip) - 1; i >= 0; i-- {
		ip[i]++
		if ip[i] != 0 {
			break
		}
	}

	return ip
}

// IPToCIDR IP地址转CIDR
func IPToCIDR(ip string, mask int) string {
	return fmt.Sprintf("%s/%d", ip, mask)
}

// CIDRToIPRange CIDR转IP范围
func CIDRToIPRange(cidr string) (string, string, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", "", err
	}

	ipv4 := ip.To4()
	if ipv4 == nil {
		return "", "", fmt.Errorf("only IPv4 addresses supported")
	}

	mask := binary.BigEndian.Uint32(ipnet.Mask)
	start := binary.BigEndian.Uint32(ipv4) & mask

	// 计算结束地址
	ones, _ := ipnet.Mask.Size()
	var end uint32
	if ones == 32 {
		end = start
	} else {
		end = start | (1<<(32-ones) - 1)
	}

	return LongToIP(start), LongToIP(end), nil
}

// GetMACAddress 获取MAC地址
func GetMACAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			return iface.HardwareAddr.String(), nil
		}
	}

	return "", fmt.Errorf("no active interface found")
}

// GetMACAddresses 获取所有MAC地址
func GetMACAddresses() (map[string]string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	macs := make(map[string]string)
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			macs[iface.Name] = iface.HardwareAddr.String()
		}
	}

	return macs, nil
}

// IsValidMAC 检查是否为有效MAC地址
func IsValidMAC(mac string) bool {
	_, err := net.ParseMAC(mac)
	return err == nil
}

// NormalizeMAC 标准化MAC地址
func NormalizeMAC(mac string) string {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return mac
	}

	return hw.String()
}

// FormatMAC 格式化MAC地址
func FormatMAC(mac string, separator string) string {
	hw, err := net.ParseMAC(mac)
	if err != nil {
		return mac
	}

	if len(hw) != 6 {
		return mac
	}

	return fmt.Sprintf("%02x%s%02x%s%02x%s%02x%s%02x%s%02x",
		hw[0], separator, hw[1], separator, hw[2], separator,
		hw[3], separator, hw[4], separator, hw[5])
}

// GetHostname 获取主机名
func GetHostname() (string, error) {
	return os.Hostname()
}

// IsPortOpen 检查端口是否开放
func IsPortOpen(host string, port int) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// IsPortOpenWithTimeout 检查端口是否开放（带超时）
func IsPortOpenWithTimeout(host string, port int, timeout time.Duration) bool {
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetAvailablePort 获取可用端口
func GetAvailablePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()

	return listener.Addr().(*net.TCPAddr).Port, nil
}

// GetAvailablePortRange 获取端口范围内的可用端口
func GetAvailablePortRange(min, max int) (int, error) {
	for port := min; port <= max; port++ {
		address := fmt.Sprintf(":%d", port)
		listener, err := net.Listen("tcp", address)
		if err == nil {
			listener.Close()
			return port, nil
		}
	}
	return 0, fmt.Errorf("no available port in range %d-%d", min, max)
}

// PortIsInUse 检查端口是否被占用
func PortIsInUse(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return true
	}
	listener.Close()
	return false
}

// ResolveDNS 解析DNS
func ResolveDNS(hostname string) ([]string, error) {
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, err
	}

	result := make([]string, 0, len(ips))
	for _, ip := range ips {
		result = append(result, ip.String())
	}

	return result, nil
}

// ReverseDNS 反向DNS查询
func ReverseDNS(ip string) (string, error) {
	names, err := net.LookupAddr(ip)
	if err != nil {
		return "", err
	}

	if len(names) == 0 {
		return "", fmt.Errorf("no hostname found for IP: %s", ip)
	}

	return names[0], nil
}

// IsValidDomain 检查是否为有效域名
func IsValidDomain(domain string) bool {
	if domain == "" || len(domain) > 253 {
		return false
	}

	// 简单的域名验证
	parts := strings.Split(domain, ".")
	if len(parts) < 2 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}

		for _, c := range part {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-') {
				return false
			}
		}
	}

	return true
}

// IsValidURL 检查是否为有效URL
func IsValidURL(urlStr string) bool {
	_, err := url.Parse(urlStr)
	return err == nil
}

// GetURLScheme 获取URL协议
func GetURLScheme(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Scheme
}

// GetURLHost 获取URL主机
func GetURLHost(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Host
}

// GetURLPath 获取URL路径
func GetURLPath(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Path
}

// GetURLQuery 获取URL查询参数
func GetURLQuery(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.RawQuery
}

// BuildURL 构建URL
func BuildURL(scheme, host, path string) string {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}
	return u.String()
}

// BuildURLWithQuery 构建带查询参数的URL
func BuildURLWithQuery(scheme, host, path string, query map[string]string) string {
	u := url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}

	q := u.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	return u.String()
}

// Ping 测试网络连通性
func Ping(host string, timeout time.Duration) error {
	// 简化实现，实际应该使用ICMP
	conn, err := net.DialTimeout("tcp", host+":80", timeout)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

// TraceRoute 路由追踪（简化实现）
func TraceRoute(host string, maxTTL int) ([]string, error) {
	if maxTTL <= 0 {
		maxTTL = 30
	}

	hops := make([]string, 0, maxTTL)

	for ttl := 1; ttl <= maxTTL; ttl++ {
		// 简化实现
		hops = append(hops, fmt.Sprintf("hop %d", ttl))

		// 检查是否到达目标
		if IsPortOpen(host, 80) {
			hops = append(hops, host)
			break
		}
	}

	return hops, nil
}

// GetNetworkInterfaces 获取网络接口
func GetNetworkInterfaces() ([]net.Interface, error) {
	return net.Interfaces()
}

// GetInterfaceIPs 获取接口IP地址
func GetInterfaceIPs(interfaceName string) ([]string, error) {
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return nil, err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, err
	}

	ips := make([]string, 0, len(addrs))
	for _, addr := range addrs {
		ips = append(ips, addr.String())
	}

	return ips, nil
}

// GetExternalIP 获取外网IP
func GetExternalIP() (string, error) {
	return GetPublicIP()
}

// IPLongToIP IP长整型转IP字符串
func IPLongToIP(long uint32) string {
	return LongToIP(long)
}

// IPToIPRange IP地址转IP范围
func IPToIPRange(ip string, mask int) (string, string, error) {
	cidr := fmt.Sprintf("%s/%d", ip, mask)
	return CIDRToIPRange(cidr)
}

// IsPrivateIP 检查是否为私有IP
func IsPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, cidr := range privateRanges {
		_, ipnet, _ := net.ParseCIDR(cidr)
		if ipnet.Contains(ip) {
			return true
		}
	}

	return false
}

// IsLocalhost 检查是否为本地地址
func IsLocalhost(host string) bool {
	return host == "localhost" ||
		host == "127.0.0.1" ||
		host == "::1" ||
		strings.HasPrefix(host, "127.") ||
		IsPrivateIP(host)
}

// ParseCIDR 解析CIDR
func ParseCIDR(cidr string) (net.IP, *net.IPNet, error) {
	return net.ParseCIDR(cidr)
}

// CIDRContains 检查CIDR是否包含IP
func CIDRContains(cidr, ipStr string) (bool, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return false, err
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	return ipnet.Contains(ip), nil
}

// GetIPTypes 获取IP类型
func GetIPTypes(ipStr string) []string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return []string{}
	}

	types := make([]string, 0)

	if ip.To4() != nil {
		types = append(types, "IPv4")
	} else {
		types = append(types, "IPv6")
	}

	if IsPrivateIP(ipStr) {
		types = append(types, "Private")
	} else {
		types = append(types, "Public")
	}

	if IsLocalhost(ipStr) {
		types = append(types, "Localhost")
	}

	return types
}
