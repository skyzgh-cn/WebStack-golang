package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// timeNow 返回当前时间，便于在测试中控制时间流逝。
func timeNow() time.Time {
	return time.Now()
}

func TestSessionLifecycle(t *testing.T) {
	m := &Manager{sessions: make(map[string]*Session)}

	token, sess, err := m.Create("admin")
	if err != nil {
		t.Fatalf("Create 失败: %v", err)
	}
	if token == "" || sess.CSRFToken == "" {
		t.Fatal("会话令牌或 CSRF 令牌为空")
	}
	if sess.Username != "admin" {
		t.Fatalf("期望用户名 admin，实际 %s", sess.Username)
	}

	got := m.Get(token)
	if got == nil {
		t.Fatal("有效令牌应能查到会话")
	}
	if got.CSRFToken != sess.CSRFToken {
		t.Fatal("会话令牌应保持稳定")
	}

	// 登出后令牌立即失效
	m.Delete(token)
	if m.Get(token) != nil {
		t.Fatal("登出后会话应失效")
	}
}

func TestSessionRejectsUnknownToken(t *testing.T) {
	m := &Manager{sessions: make(map[string]*Session)}
	if m.Get("forged-token") != nil {
		t.Fatal("伪造令牌不应获得会话")
	}
	if m.Get("") != nil {
		t.Fatal("空令牌不应获得会话")
	}
}

func TestSessionIdleTimeout(t *testing.T) {
	now := timeNow()
	m := &Manager{sessions: make(map[string]*Session), now: func() time.Time { return now }}

	token, _, err := m.Create("admin")
	if err != nil {
		t.Fatalf("Create 失败: %v", err)
	}

	// 推进时间超过空闲超时
	now = now.Add(SessionIdleTimeout + time.Minute)
	if m.Get(token) != nil {
		t.Fatal("超过空闲超时后会话应失效")
	}
}

func TestSessionAbsoluteExpiry(t *testing.T) {
	now := timeNow()
	m := &Manager{sessions: make(map[string]*Session), now: func() time.Time { return now }}

	token, _, err := m.Create("admin")
	if err != nil {
		t.Fatalf("Create 失败: %v", err)
	}

	// 每次访问都刷新空闲时间，但绝对过期时间不变
	now = now.Add(SessionIdleTimeout - time.Minute)
	if m.Get(token) == nil {
		t.Fatal("空闲超时前会话应有效")
	}
	now = now.Add(SessionTTL)
	if m.Get(token) != nil {
		t.Fatal("超过最长存活时间后会话应失效")
	}
}

func TestVerifyLoginCSRF(t *testing.T) {
	rec := httptest.NewRecorder()
	token, err := IssueLoginCSRF(rec, false)
	if err != nil {
		t.Fatalf("IssueLoginCSRF 失败: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/login", nil)
	req.AddCookie(&http.Cookie{Name: LoginCSRFCookieName, Value: token})

	if !VerifyLoginCSRF(req, token) {
		t.Fatal("令牌匹配时应通过校验")
	}
	if VerifyLoginCSRF(req, "wrong-token") {
		t.Fatal("错误令牌不应通过校验")
	}
	if VerifyLoginCSRF(req, "") {
		t.Fatal("空令牌不应通过校验")
	}

	// 无 Cookie 时一律拒绝
	noCookie := httptest.NewRequest(http.MethodPost, "/admin/login", nil)
	if VerifyLoginCSRF(noCookie, token) {
		t.Fatal("缺少 Cookie 时不应通过校验")
	}
}

func TestLoginRateLimiter(t *testing.T) {
	ip, user := "10.0.0.1", "admin"
	for i := 0; i < loginMaxFailures; i++ {
		if !AllowLogin(ip, user) {
			t.Fatalf("第 %d 次尝试应被允许", i+1)
		}
		RecordLoginFailure(ip, user)
	}

	if AllowLogin(ip, user) {
		t.Fatal("超过失败阈值后应被拒绝")
	}

	// 其他来源不受影响
	if !AllowLogin("10.0.0.2", user) {
		t.Fatal("不同 IP 不应被牵连")
	}

	// 登录成功后重置
	ResetLoginFailures(ip, user)
	if !AllowLogin(ip, user) {
		t.Fatal("重置后应恢复允许")
	}
}
