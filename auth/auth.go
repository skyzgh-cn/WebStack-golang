// Package auth 提供后台会话管理与 CSRF 令牌。
//
// 会话采用服务端存储：Cookie 中仅存放随机不可猜测的令牌，用户名、CSRF 令牌
// 等敏感信息保存在服务端内存中。登出或超时后令牌立即失效，无法被伪造。
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// CookieName 存放会话令牌的 Cookie 名称。
const CookieName = "webstack_session"

// 会话有效期设置。
const (
	// SessionTTL 会话最长存活时间，超过后强制重新登录。
	SessionTTL = 2 * time.Hour
	// SessionIdleTimeout 空闲超时，超过该时间未活动则会话失效。
	SessionIdleTimeout = 30 * time.Minute
)

// Session 表示一个已登录的后台会话。
type Session struct {
	Username  string
	CSRFToken string
	CreatedAt time.Time
	ExpiresAt time.Time
	LastSeen  time.Time
}

// Manager 负责会话的创建、查询与销毁。
type Manager struct {
	mu       sync.Mutex
	sessions map[string]*Session
	now      func() time.Time
}

var defaultManager = &Manager{sessions: make(map[string]*Session), now: time.Now}

// Default 返回全局会话管理器。
func Default() *Manager {
	return defaultManager
}

// randomToken 生成密码学安全的随机令牌。
func randomToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// Create 为指定用户名创建一个新会话，并返回会话令牌与 CSRF 令牌。
func (m *Manager) Create(username string) (string, *Session, error) {
	token, err := randomToken()
	if err != nil {
		return "", nil, err
	}
	csrf, err := randomToken()
	if err != nil {
		return "", nil, err
	}

	now := m.clock()
	sess := &Session{
		Username:  username,
		CSRFToken: csrf,
		CreatedAt: now,
		ExpiresAt: now.Add(SessionTTL),
		LastSeen:  now,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[token] = sess
	return token, sess, nil
}

// clock 返回当前时间，测试中可替换 now 字段以控制时间流逝。
func (m *Manager) clock() time.Time {
	if m.now == nil {
		return time.Now()
	}
	return m.now()
}

// Get 校验令牌并返回对应会话；令牌无效或已过期时返回 nil。
func (m *Manager) Get(token string) *Session {
	if token == "" {
		return nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	sess, ok := m.sessions[token]
	if !ok {
		return nil
	}

	now := m.clock()
	if now.After(sess.ExpiresAt) || now.Sub(sess.LastSeen) > SessionIdleTimeout {
		delete(m.sessions, token)
		return nil
	}

	// 滑动刷新空闲超时
	sess.LastSeen = now
	return sess
}

// Delete 立即销毁指定令牌对应的会话。
func (m *Manager) Delete(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
}

// Cleanup 清理所有已过期的会话，由后台定时调用。
func (m *Manager) Cleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.clock()
	for token, sess := range m.sessions {
		if now.After(sess.ExpiresAt) || now.Sub(sess.LastSeen) > SessionIdleTimeout {
			delete(m.sessions, token)
		}
	}
}

// Count 返回当前有效会话数，仅用于监控与测试。
func (m *Manager) Count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sessions)
}

// StartCleanupLoop 周期性清理过期会话，直至 stop 被关闭。
func (m *Manager) StartCleanupLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				m.Cleanup()
			case <-stop:
				return
			}
		}
	}()
}

// SetSessionCookie 将会话令牌写入 Cookie。secure 为 true 时仅通过 HTTPS 传输。
func SetSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   int(SessionTTL / time.Second),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// ClearSessionCookie 清除客户端会话 Cookie。
func ClearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

// SessionFromRequest 从请求中解析并校验会话令牌。
func SessionFromRequest(r *http.Request) *Session {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		return nil
	}
	return Default().Get(cookie.Value)
}
