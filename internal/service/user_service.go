package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"products/internal/domain"
	"products/internal/repository"
)

// UserService implements the user service interface
type UserService struct {
	userRepo       *repository.UserRepository
	sessionService *SessionService
	jwtSecret      string
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository, sessionService *SessionService, jwtSecret string) *UserService {
	return &UserService{
		userRepo:       userRepo,
		sessionService: sessionService,
		jwtSecret:      jwtSecret,
	}
}

// Register creates a new user account
func (s *UserService) Register(ctx context.Context, user *domain.User) error {
	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err == nil && existingUser != nil {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.ID = uuid.New()
	user.Password = string(hashedPassword)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	return s.userRepo.Create(ctx, user)
}

// Login authenticates a user and returns access and refresh tokens
func (s *UserService) Login(ctx context.Context, email, password, ipAddress, userAgent string) (*domain.LoginResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	session, err := s.sessionService.CreateSession(ctx, user.ID.String(), user.Email, ipAddress, userAgent, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	accessToken, err := s.generateAccessToken(user, session.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.generateRefreshToken(user, session.ID)
	if err != nil {
		return nil, err
	}

	user.Password = ""

	response := &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
		ExpiresIn:    3600, // 1 hour
	}

	return response, nil
}

// RefreshToken generates new access and refresh tokens
func (s *UserService) RefreshToken(ctx context.Context, refreshToken string) (*domain.RefreshTokenResponse, error) {
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("invalid user ID in token")
	}

	sessionID, ok := claims["session_id"].(string)
	if !ok {
		return nil, errors.New("invalid session ID in token")
	}

	isValid, err := s.sessionService.IsSessionValid(ctx, sessionID)
	if err != nil || !isValid {
		return nil, errors.New("session expired or invalid")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	accessToken, err := s.generateAccessToken(user, sessionID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.generateRefreshToken(user, sessionID)
	if err != nil {
		return nil, err
	}

	err = s.sessionService.RefreshSession(ctx, sessionID, 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh session: %w", err)
	}

	return &domain.RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    3600, // 1 hour
	}, nil
}

// Logout invalidates a user session
func (s *UserService) Logout(ctx context.Context, sessionID string) error {

	return s.sessionService.DeleteSession(ctx, sessionID)
}

// LogoutAll invalidates all sessions for a user
func (s *UserService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if err := s.BlacklistAllUserSessions(ctx, userID); err != nil {
		return fmt.Errorf("failed to blacklist user sessions: %w", err)
	}

	return s.sessionService.DeleteUserSessions(ctx, userID.String())
}

// BlacklistAllUserSessions blacklists all sessions for a specific user
func (s *UserService) BlacklistAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	sessions, err := s.sessionService.GetUserSessions(ctx, userID.String())
	if err != nil {
		return fmt.Errorf("failed to get user sessions: %w", err)
	}

	for _, session := range sessions {
		userBlacklistKey := fmt.Sprintf("user_blacklist:%s:%s", userID.String(), session.ID)

		if err := s.sessionService.cacheService.Set(ctx, userBlacklistKey, true, 24*time.Hour); err != nil {
			return fmt.Errorf("failed to blacklist session %s: %w", session.ID, err)
		}
	}

	return nil
}

// ValidateSession checks if a session is still valid
func (s *UserService) ValidateSession(ctx context.Context, sessionID string) (bool, error) {
	return s.sessionService.IsSessionValid(ctx, sessionID)
}

// IsTokenBlacklisted checks if a token has been blacklisted
func (s *UserService) IsTokenBlacklisted(ctx context.Context, token string) (bool, error) {
	tokenHash := s.hashToken(token)
	blacklistKey := fmt.Sprintf("blacklist:%s", tokenHash)

	exists, err := s.sessionService.cacheService.Exists(ctx, blacklistKey)
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	return exists, nil
}

// IsUserSessionBlacklisted checks if a user's session has been blacklisted by logout all
func (s *UserService) IsUserSessionBlacklisted(ctx context.Context, userID uuid.UUID, sessionID string) (bool, error) {
	userBlacklistKey := fmt.Sprintf("user_blacklist:%s:%s", userID.String(), sessionID)
	exists, err := s.sessionService.cacheService.Exists(ctx, userBlacklistKey)
	if err != nil {
		return false, fmt.Errorf("failed to check user session blacklist: %w", err)
	}
	return exists, nil
}

// BlacklistToken adds a token to the blacklist
func (s *UserService) BlacklistToken(ctx context.Context, token string) error {
	tokenHash := s.hashToken(token)
	blacklistKey := fmt.Sprintf("blacklist:%s", tokenHash)

	return s.sessionService.cacheService.Set(ctx, blacklistKey, true, 24*time.Hour)
}

// hashToken creates a proper cryptographic hash of the token for blacklisting
func (s *UserService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// GetUserSessions returns all active sessions for a user
func (s *UserService) GetUserSessions(ctx context.Context, userID uuid.UUID) (*domain.UserSessionsResponse, error) {

	count, err := s.sessionService.GetActiveSessionsCount(ctx, userID.String())
	if err != nil {
		return nil, err
	}

	return &domain.UserSessionsResponse{
		ActiveSessions: []domain.SessionInfo{}, // Would need to implement this
		TotalSessions:  count,
	}, nil
}

// GetByID retrieves a user by ID
func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// generateAccessToken generates a short-lived access token
func (s *UserService) generateAccessToken(user *domain.User, sessionID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"session_id": sessionID,
		"exp":        time.Now().Add(time.Hour).Unix(), // 1 hour
		"type":       "access",
	})

	return token.SignedString([]byte(s.jwtSecret))
}

// generateRefreshToken generates a long-lived refresh token
func (s *UserService) generateRefreshToken(user *domain.User, sessionID string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"session_id": sessionID,
		"exp":        time.Now().Add(7 * 24 * time.Hour).Unix(), // 7 days
		"type":       "refresh",
	})

	return token.SignedString([]byte(s.jwtSecret))
}
