package auth

import (
	"crypto/subtle"
	"net/http"
	"sync"
	"time"
)

// LoginCSRFCookieName 存放登录 CSRF 令牌的 Cookie 名称。
const LoginCSRFCookieName = "webstack_login_csrf"

// LoginCSRFField 登录表单中携带 CSRF 令牌的字段名。
const LoginCSRFField = "_csrf"

// IssueLoginCSRF 生成一个新的登录 CSRF 令牌并写入 Cookie，同时返回令牌值。
//
// 登录发生在建立会话之前，无法使用会话内令牌，因此采用"双提交 Cookie"模式：
// 令牌同时存在于 Cookie 与页面隐藏域中，提交时必须一致。
func IssueLoginCSRF(w http.ResponseWriter, secure bool) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     LoginCSRFCookieName,
		Value:    token,
		Path:     "/admin/login",
		MaxAge:   int(15 * time.Minute / time.Second),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	})

	return token, nil
}

// VerifyLoginCSRF 校验登录请求携带的 CSRF 令牌是否与 Cookie 中的值一致。
func VerifyLoginCSRF(r *http.Request, submitted string) bool {
	cookie, err := r.Cookie(LoginCSRFCookieName)
	if err != nil || cookie.Value == "" {
		return false
	}
	if submitted == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(submitted)) == 1
}

// 登录失败次数限制，防止暴力破解。

// 登录失败阈值与锁定时长。
const (
	loginMaxFailures   = 5
	loginLockoutWindow = 15 * time.Minute
)

// loginLimiter 以"客户端 IP + 用户名"为维度记录登录失败次数。
type loginLimiter struct {
	mu       sync.Mutex
	failures map[string][]time.Time
}

var limiter = &loginLimiter{failures: make(map[string][]time.Time)}

// key 生成限流键。
func loginKey(ip, username string) string {
	return ip + "|" + username
}

// Allow 判断当前是否允许继续尝试登录。
func (l *loginLimiter) Allow(ip, username string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	records := l.prune(now, loginKey(ip, username))
	return len(records) < loginMaxFailures
}

// Fail 记录一次登录失败。
func (l *loginLimiter) Fail(ip, username string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	key := loginKey(ip, username)
	l.failures[key] = append(l.prune(now, key), now)
}

// Reset 在登录成功后清除失败记录。
func (l *loginLimiter) Reset(ip, username string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.failures, loginKey(ip, username))
}

// prune 清理窗口期外的失败记录。
func (l *loginLimiter) prune(now time.Time, key string) []time.Time {
	cutoff := now.Add(-loginLockoutWindow)
	records := l.failures[key]
	kept := records[:0]
	for _, t := range records {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		delete(l.failures, key)
		return nil
	}
	l.failures[key] = kept
	return kept
}

// AllowLogin 判断是否允许该客户端继续尝试登录。
func AllowLogin(ip, username string) bool {
	return limiter.Allow(ip, username)
}

// RecordLoginFailure 记录一次登录失败。
func RecordLoginFailure(ip, username string) {
	limiter.Fail(ip, username)
}

// ResetLoginFailures 登录成功后清除失败计数。
func ResetLoginFailures(ip, username string) {
	limiter.Reset(ip, username)
}
