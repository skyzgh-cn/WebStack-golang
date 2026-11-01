package admin

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// maxHTMLBodySize 限制抓取的 HTML 内容最大字节数，防止内存耗尽。
const maxHTMLBodySize = 2 << 20 // 2MB

// maxRedirects 允许跟随的最大重定向次数。
const maxRedirects = 5

// SSRF 防护：抓取外部网页时阻止访问内网 / 本机 / 云元数据地址，
// 并通过自定义 DialContext 在**建立连接的瞬间**再次校验目标 IP，
// 以防 DNS 重绑定（DNS rebinding）绕过前置校验。

// allowedFetchPorts 允许抓取的端口，仅开放 Web 常用端口。
var allowedFetchPorts = map[string]bool{
	"80":  true,
	"443": true,
}

// blockedHostnames 明确禁止访问的主机名。
var blockedHostnames = map[string]bool{
	"metadata.google.internal": true,
	"metadata.goog":            true,
	"metadata":                 true,
	"instance-data":            true,
}

// errBlockedTarget 表示目标地址属于禁止访问的范围。
var errBlockedTarget = errors.New("不允许访问该目标地址")

// isBlockedIP 判断 IP 是否属于内网、本机、链路本地或保留网段。
func isBlockedIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	if ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsUnspecified() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() {
		return true
	}

	// 云厂商元数据地址 169.254.169.254 已被 LinkLocalUnicast 覆盖，
	// 这里额外覆盖 IPv4 转换后的 100.64.0.0/10（运营商级 NAT）与 192.0.0.0/24 等保留段
	if v4 := ip.To4(); v4 != nil {
		switch {
		case v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127: // 100.64.0.0/10
			return true
		case v4[0] == 192 && v4[1] == 0 && v4[2] == 0: // 192.0.0.0/24
			return true
		case v4[0] == 0: // 0.0.0.0/8
			return true
		case v4[0] >= 240: // 240.0.0.0/4 保留
			return true
		}
	}

	return false
}

// validateTargetURL 校验 URL 协议、主机与端口是否合法。
func validateTargetURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return nil, fmt.Errorf("无效的 URL 格式: %w", err)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, errors.New("仅支持 http/https 协议")
	}

	host := parsed.Hostname()
	if host == "" {
		return nil, errors.New("URL 缺少主机名")
	}

	if blockedHostnames[strings.ToLower(host)] {
		return nil, errBlockedTarget
	}

	// 字面量 IP 直接判定，避免依赖后续 DNS 解析
	if ip := net.ParseIP(host); ip != nil && isBlockedIP(ip) {
		return nil, errBlockedTarget
	}

	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	if !allowedFetchPorts[port] {
		return nil, fmt.Errorf("不允许访问端口 %s", port)
	}

	return parsed, nil
}

// resolveAndCheckHost 解析主机名并确认其所有解析结果都不在内网。
func resolveAndCheckHost(ctx context.Context, host string) error {
	if ip := net.ParseIP(host); ip != nil {
		if isBlockedIP(ip) {
			return errBlockedTarget
		}
		return nil
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("域名解析失败: %w", err)
	}
	if len(ips) == 0 {
		return errors.New("域名未解析到任何地址")
	}

	for _, ip := range ips {
		if isBlockedIP(ip.IP) {
			return errBlockedTarget
		}
	}

	return nil
}

// safeDialContext 在建立 TCP 连接前校验目标 IP，阻断 DNS 重绑定攻击。
func safeDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	if err := resolveAndCheckHost(ctx, host); err != nil {
		return nil, err
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second, KeepAlive: 30 * time.Second}
	conn, err := dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	// 连接建立后再校验一次实际对端 IP
	if tcpAddr, ok := conn.RemoteAddr().(*net.TCPAddr); ok {
		if isBlockedIP(tcpAddr.IP) {
			conn.Close()
			return nil, errBlockedTarget
		}
	}

	return conn, nil
}

// newSafeHTTPClient 构造用于抓取外部网页的客户端。
//
// 安全要点：
//   - 使用自定义 DialContext，在建连瞬间校验目标 IP，阻断 DNS 重绑定；
//   - 每次重定向都重新校验目标 URL，阻断通过 302 跳转访问内网；
//   - 限制超时与响应体大小。
func newSafeHTTPClient() *http.Client {
	transport := &http.Transport{
		DialContext:           safeDialContext,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		IdleConnTimeout:       30 * time.Second,
		DisableCompression:    false,
	}

	return &http.Client{
		Timeout:   15 * time.Second,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return errors.New("重定向次数过多")
			}
			// 每一跳都重新校验协议、主机与端口
			if _, err := validateTargetURL(req.URL.String()); err != nil {
				return err
			}
			return resolveAndCheckHost(req.Context(), req.URL.Hostname())
		},
	}
}
