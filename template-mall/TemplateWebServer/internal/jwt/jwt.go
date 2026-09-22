// Package jwt 提供 JWT access token 解析功能。
// 与 TemplateOrderServer/internal/jwt 保持一致的 Claims 结构，
// 确保能正确解析由 TemplateOrderServer 签发的 token。
package jwt

import (
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenInvalid = errors.New("token invalid")
)

// Claims JWT payload（与 TemplateOrderServer 保持一致）
type Claims struct {
	UserID       uint64 `json:"user_id"`
	Username     string `json:"username"`
	MemberStatus int8   `json:"member_status"`
	jwt.RegisteredClaims
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
