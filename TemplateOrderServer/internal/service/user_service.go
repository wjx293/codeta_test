package service

import (
	"errors"
	"strings"
	"time"

	"template-mall/TemplateOrderServer/internal/model"
	"template-mall/TemplateOrderServer/internal/repository"

	jwtutil "template-mall/TemplateOrderServer/internal/jwt"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUsernameExists     = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrTokenRevoked       = errors.New("token revoked")
	ErrTokenExpired       = errors.New("token expired")
)

// UserService 用户服务
type UserService struct {
	userRepo         *repository.UserRepo
	refreshTokenRepo *repository.RefreshTokenRepo
	jwtSecret        string
	accessTTL        time.Duration
	refreshTTL       time.Duration
}

// NewUserService 创建用户服务实例。
func NewUserService(
	userRepo *repository.UserRepo,
	refreshTokenRepo *repository.RefreshTokenRepo,
	jwtSecret string,
	accessTTL, refreshTTL time.Duration,
) *UserService {
	return &UserService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtSecret:        jwtSecret,
		accessTTL:        accessTTL,
		refreshTTL:       refreshTTL,
	}
}

// RegisterResponse 注册/登录/刷新响应
type RegisterResponse struct {
	UserID       uint64
	Username     string
	MemberStatus int8
	AccessToken  string
	RefreshToken string
}

// Register 用户注册：用户名转小写、bcrypt cost 10、捕获 uk_username 冲突。
func (s *UserService) Register(username, password string) (*RegisterResponse, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &model.TosUser{
		Username:  username,
		Password:  string(hash),
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.Create(user); err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			return nil, ErrUsernameExists
		}
		return nil, err
	}

	return s.generateTokens(user)
}

// Login 用户登录：校验密码、签发 access(2h) + refresh(7d)。
func (s *UserService) Login(username, password string) (*RegisterResponse, error) {
	username = strings.ToLower(strings.TrimSpace(username))

	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(user)
}

// RefreshToken 刷新 access token：校验 refresh 存在且未吊销且未过期、签发新对、旧 refresh 标记 revoked。
func (s *UserService) RefreshToken(refreshToken string) (*RegisterResponse, error) {
	tokenHash := jwtutil.HashToken(refreshToken)

	record, err := s.refreshTokenRepo.FindByTokenHash(tokenHash)
	if err != nil {
		return nil, errors.New("token invalid")
	}

	if record.Revoked == 1 {
		return nil, ErrTokenRevoked
	}

	if time.Now().After(record.ExpiresAt) {
		return nil, ErrTokenExpired
	}

	user, err := s.userRepo.FindByID(record.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// 吊销旧 refresh token
	_ = s.refreshTokenRepo.RevokeByTokenHash(tokenHash)

	// 签发新 token 对
	return s.generateTokens(user)
}

// RevokeRefreshTokens 吊销用户所有未吊销的 refresh token（FR-006 登出）。
func (s *UserService) RevokeRefreshTokens(userID uint64) error {
	return s.refreshTokenRepo.RevokeByUserID(userID)
}

// generateTokens 签发 access + refresh token 对。
func (s *UserService) generateTokens(user *model.TosUser) (*RegisterResponse, error) {
	accessToken, err := jwtutil.GenerateAccessToken(s.jwtSecret, user.ID, user.Username, user.MemberStatus, s.accessTTL)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwtutil.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}
	tokenHash := jwtutil.HashToken(refreshToken)

	record := &model.TosRefreshToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.refreshTTL),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(record); err != nil {
		return nil, err
	}

	return &RegisterResponse{
		UserID:       user.ID,
		Username:     user.Username,
		MemberStatus: user.MemberStatus,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// ListUsers 分页查询用户，支持 username 模糊搜索（B 端用户检索）。
func (s *UserService) ListUsers(username string, offset, limit int) ([]model.TosUser, int64, error) {
	return s.userRepo.ListUsers(username, offset, limit)
}
