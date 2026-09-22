// Package jwt 提供 JWT access token 与 refresh token 的签发与验证。
package jwt

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

// Claims JWT payload
type Claims struct {
	UserID       uint64 `json:"user_id"`
	Username     string `json:"username"`
	MemberStatus int8   `json:"member_status"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 签发 access token（2h）。
func GenerateAccessToken(secret string, userID uint64, username string, memberStatus int8, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:       userID,
		Username:     username,
		MemberStatus: memberStatus,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseAccessToken 解析并验证 access token。
func ParseAccessToken(secret, tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrTokenExpired
		}
		return nil, ErrTokenInvalid
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrTokenInvalid
	}
	return claims, nil
}

// GenerateRefreshToken 生成 refresh token（32 字节密码学安全随机数，hex 编码 64 字符）。
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashToken 对 token 做 SHA256 哈希。
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}