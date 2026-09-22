package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"template-mall/TemplateWebServer/internal/jwt"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-key"

// generateTestToken 生成测试用 access token。
func generateTestToken(t *testing.T, secret string, userID uint64, username string, memberStatus int8, exp time.Time) string {
	claims := &jwt.Claims{
		UserID:       userID,
		Username:     username,
		MemberStatus: memberStatus,
		RegisteredClaims: jwtpkg.RegisteredClaims{
			ExpiresAt: jwtpkg.NewNumericDate(exp),
			IssuedAt:  jwtpkg.NewNumericDate(time.Now()),
		},
	}
	token := jwtpkg.NewWithClaims(jwtpkg.SigningMethodHS256, claims)
	str, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return str
}

func setupRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(JWTAuth(secret))
	r.GET("/test", func(c *gin.Context) {
		uid, _ := GetUserID(c)
		c.JSON(http.StatusOK, gin.H{"user_id": uid})
	})
	return r
}

func TestJWTAuth_Success(t *testing.T) {
	r := setupRouter(testSecret)
	token := generateTestToken(t, testSecret, 42, "alice", 1, time.Now().Add(time.Hour))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !contains(w.Body.String(), `"user_id":42`) {
		t.Fatalf("expected user_id 42 in body, got %s", w.Body.String())
	}
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	r := setupRouter(testSecret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	r := setupRouter(testSecret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic abc123")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_ExpiredToken(t *testing.T) {
	r := setupRouter(testSecret)
	token := generateTestToken(t, testSecret, 42, "alice", 1, time.Now().Add(-time.Hour))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	if !contains(w.Body.String(), "expired") {
		t.Fatalf("expected 'expired' in body, got %s", w.Body.String())
	}
}

func TestJWTAuth_InvalidSignature(t *testing.T) {
	r := setupRouter(testSecret)
	token := generateTestToken(t, "wrong-secret", 42, "alice", 1, time.Now().Add(time.Hour))

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestJWTAuth_EmptyToken(t *testing.T) {
	r := setupRouter(testSecret)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOf(s, substr) >= 0))
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}