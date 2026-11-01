package admin

import (
	"context"
	"net"
	"testing"
)

func TestIsBlockedIP(t *testing.T) {
	blocked := []string{
		"127.0.0.1",
		"127.0.0.53",
		"0.0.0.0",
		"10.0.0.1",
		"172.16.0.1",
		"172.31.255.255",
		"192.168.1.1",
		"169.254.169.254", // 云元数据
		"100.64.0.1",      // 运营商级 NAT
		"0.0.0.0",
		"224.0.0.1", // 组播
		"::1",       // IPv6 回环
		"fc00::1",   // IPv6 ULA
		"fe80::1",   // IPv6 链路本地
	}

	for _, s := range blocked {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("测试用例 %s 不是合法 IP", s)
		}
		if !isBlockedIP(ip) {
			t.Errorf("IP %s 应被拦截", s)
		}
	}

	allowed := []string{
		"8.8.8.8",
		"1.1.1.1",
		"93.184.216.34",
		"2001:4860:4860::8888",
	}

	for _, s := range allowed {
		ip := net.ParseIP(s)
		if ip == nil {
			t.Fatalf("测试用例 %s 不是合法 IP", s)
		}
		if isBlockedIP(ip) {
			t.Errorf("公网 IP %s 不应被拦截", s)
		}
	}
}

func TestValidateTargetURL(t *testing.T) {
	// 非 http/https 协议应被拒绝
	if _, err := validateTargetURL("file:///etc/passwd"); err == nil {
		t.Error("file 协议应被拒绝")
	}
	if _, err := validateTargetURL("gopher://127.0.0.1:6379/_x"); err == nil {
		t.Error("gopher 协议应被拒绝")
	}

	// 内网地址应被拒绝
	if _, err := validateTargetURL("http://127.0.0.1/"); err == nil {
		t.Error("回环地址应被拒绝")
	}
	if _, err := validateTargetURL("http://169.254.169.254/latest/meta-data/"); err == nil {
		t.Error("云元数据地址应被拒绝")
	}
	if _, err := validateTargetURL("http://192.168.1.1/"); err == nil {
		t.Error("内网地址应被拒绝")
	}

	// 非白名单端口应被拒绝
	if _, err := validateTargetURL("http://example.com:6379/"); err == nil {
		t.Error("非白名单端口应被拒绝")
	}

	// 缺失主机名应被拒绝
	if _, err := validateTargetURL("http:///path"); err == nil {
		t.Error("缺少主机名应被拒绝")
	}

	// 正常公网 URL 应通过
	if _, err := validateTargetURL("https://example.com/path?q=1"); err != nil {
		t.Errorf("正常公网 URL 应通过校验: %v", err)
	}

	// 80/443 之外的端口按设计拒绝
	if _, err := validateTargetURL("http://example.com:8080/"); err == nil {
		t.Error("白名单之外的端口应被拒绝")
	}
}

func TestResolveAndCheckHost(t *testing.T) {
	ctx := context.Background()

	// 字面量内网 IP
	if err := resolveAndCheckHost(ctx, "127.0.0.1"); err == nil {
		t.Error("回环 IP 应被拒绝")
	}
	if err := resolveAndCheckHost(ctx, "169.254.169.254"); err == nil {
		t.Error("云元数据 IP 应被拒绝")
	}

	// 公网 IP 应通过
	if err := resolveAndCheckHost(ctx, "8.8.8.8"); err != nil {
		t.Errorf("公网 IP 应通过: %v", err)
	}
}
