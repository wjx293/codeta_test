package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// Session 服务端会话。
type Session struct {
	ID         string
	UserID     string
	Username   string
	Avatar     string
	CreatedAt  time.Time
	ExpiresAt  time.Time
}

// SessionStore 服务端会话存储（内存版，作业范围足够）。
// 生产环境可换 Redis 实现，接口保持一致。
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	ttl      time.Duration
}

// NewSessionStore 创建会话存储。
func NewSessionStore(ttl time.Duration) *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
		ttl:      ttl,
	}
}

// Create 创建新会话，返回 session id（写入 Cookie）。
func (s *SessionStore) Create(userID, username, avatar string) string {
	sid := randomID(32)
	now := time.Now()
	session := &Session{
		ID:        sid,
		UserID:    userID,
		Username:  username,
		Avatar:    avatar,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}

	s.mu.Lock()
	s.sessions[sid] = session
	s.mu.Unlock()

	return sid
}

// Get 获取会话，过期返回 nil。
func (s *SessionStore) Get(sid string) *Session {
	s.mu.RLock()
	session := s.sessions[sid]
	s.mu.RUnlock()

	if session == nil {
		return nil
	}
	if time.Now().After(session.ExpiresAt) {
		s.Delete(sid)
		return nil
	}
	return session
}

// Delete 删除会话（幂等）。
func (s *SessionStore) Delete(sid string) {
	s.mu.Lock()
	delete(s.sessions, sid)
	s.mu.Unlock()
}

// randomID 生成随机 hex 字符串。
func randomID(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// 退化到时间戳（极端情况）
		return time.Now().Format("20060102150405000000000")
	}
	return hex.EncodeToString(b)
}