package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Session represents a user session
type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	IsActive  bool      `json:"is_active"`
}

// SessionService manages user sessions
type SessionService struct {
	cacheService *CacheService
}

// NewSessionService creates a new session service
func NewSessionService(cacheService *CacheService) *SessionService {
	return &SessionService{
		cacheService: cacheService,
	}
}

// CreateSession creates a new user session
func (s *SessionService) CreateSession(ctx context.Context, userID, email, ipAddress, userAgent string, duration time.Duration) (*Session, error) {
	sessionID := uuid.New().String()
	now := time.Now()

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		Email:     email,
		CreatedAt: now,
		ExpiresAt: now.Add(duration),
		IPAddress: ipAddress,
		UserAgent: userAgent,
		IsActive:  true,
	}

	key := fmt.Sprintf("session:%s", sessionID)
	err := s.cacheService.Set(ctx, key, session, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)
	err = s.cacheService.Set(ctx, userSessionsKey, sessionID, duration)
	if err != nil {
		return nil, fmt.Errorf("failed to store user session index: %w", err)
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	key := fmt.Sprintf("session:%s", sessionID)
	var session Session

	err := s.cacheService.Get(ctx, key, &session)
	if err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}

	if time.Now().After(session.ExpiresAt) {
		s.DeleteSession(ctx, sessionID)
		return nil, fmt.Errorf("session expired")
	}

	return &session, nil
}

// DeleteSession removes a session
func (s *SessionService) DeleteSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)

	var session Session
	err := s.cacheService.Get(ctx, key, &session)
	if err == nil {
		userSessionsKey := fmt.Sprintf("user_sessions:%s", session.UserID)
		s.cacheService.Delete(ctx, userSessionsKey)
	}

	return s.cacheService.Delete(ctx, key)
}

// DeleteUserSessions removes all sessions for a specific user
func (s *SessionService) DeleteUserSessions(ctx context.Context, userID string) error {
	pattern := fmt.Sprintf("session:*")
	keys, err := s.cacheService.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get session keys: %w", err)
	}

	for _, key := range keys {
		var session Session
		if err := s.cacheService.Get(ctx, key, &session); err == nil {
			if session.UserID == userID {
				s.cacheService.Delete(ctx, key)
			}
		}
	}

	userSessionsKey := fmt.Sprintf("user_sessions:%s", userID)
	return s.cacheService.Delete(ctx, userSessionsKey)
}

// RefreshSession extends a session's expiration time
func (s *SessionService) RefreshSession(ctx context.Context, sessionID string, duration time.Duration) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ExpiresAt = time.Now().Add(duration)

	key := fmt.Sprintf("session:%s", sessionID)
	return s.cacheService.Set(ctx, key, session, duration)
}

// IsSessionValid checks if a session is valid and active
func (s *SessionService) IsSessionValid(ctx context.Context, sessionID string) (bool, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return false, nil
	}

	return session.IsActive && time.Now().Before(session.ExpiresAt), nil
}

// GetActiveSessionsCount returns the number of active sessions for a user
func (s *SessionService) GetActiveSessionsCount(ctx context.Context, userID string) (int64, error) {
	pattern := fmt.Sprintf("session:*")
	keys, err := s.cacheService.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get session keys: %w", err)
	}

	count := int64(0)
	for _, key := range keys {
		var session Session
		if err := s.cacheService.Get(ctx, key, &session); err == nil {
			if session.UserID == userID && session.IsActive && time.Now().Before(session.ExpiresAt) {
				count++
			}
		}
	}

	return count, nil
}

// GetUserSessions returns all active sessions for a user
func (s *SessionService) GetUserSessions(ctx context.Context, userID string) ([]Session, error) {
	pattern := fmt.Sprintf("session:*")
	keys, err := s.cacheService.Client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %w", err)
	}

	var sessions []Session
	for _, key := range keys {
		var session Session
		if err := s.cacheService.Get(ctx, key, &session); err == nil {
			if session.UserID == userID && session.IsActive && time.Now().Before(session.ExpiresAt) {
				sessions = append(sessions, session)
			}
		}
	}

	return sessions, nil
}
