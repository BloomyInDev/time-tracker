package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type session struct {
	UserID    int64
	ExpiresAt time.Time
}

// SessionStore holds active sessions in memory. Sessions are lost on
// server restart, which is an accepted tradeoff for this internal app in
// exchange for instant, DB-free revocation (just delete the map entry).
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]session
	ttl      time.Duration
}

func NewSessionStore(ttl time.Duration) *SessionStore {
	s := &SessionStore{
		sessions: make(map[string]session),
		ttl:      ttl,
	}
	go s.cleanupLoop()
	return s
}

func (s *SessionStore) Create(userID int64) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.sessions[token] = session{UserID: userID, ExpiresAt: time.Now().Add(s.ttl)}
	s.mu.Unlock()

	return token, nil
}

// Get returns the session's user id and slides its expiry forward, so an
// actively used session never expires while an idle one does after ttl.
func (s *SessionStore) Get(token string) (int64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.sessions[token]
	if !ok || time.Now().After(sess.ExpiresAt) {
		delete(s.sessions, token)
		return 0, false
	}

	sess.ExpiresAt = time.Now().Add(s.ttl)
	s.sessions[token] = sess
	return sess.UserID, true
}

func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

func (s *SessionStore) cleanupLoop() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for token, sess := range s.sessions {
			if now.After(sess.ExpiresAt) {
				delete(s.sessions, token)
			}
		}
		s.mu.Unlock()
	}
}

func randomToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
