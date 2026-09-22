package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"template-mall/TemplateAdminWebServer/internal/auth"

	"github.com/gin-gonic/gin"
)

func setupSessionRouter(store *auth.SessionStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SessionAuth(store))
	r.GET("/test", func(c *gin.Context) {
		s, _ := GetSession(c)
		c.JSON(http.StatusOK, gin.H{"user_id": s.UserID})
	})
	return r
}

func TestSessionAuth_Success(t *testing.T) {
	store := auth.NewSessionStore(time.Hour)
	sid := store.Create("10001", "alice", "")

	r := setupSessionRouter(store)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sid})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if !containsStr(w.Body.String(), `"user_id":"10001"`) {
		t.Fatalf("expected user_id 10001 in body, got %s", w.Body.String())
	}
}

func TestSessionAuth_MissingCookie(t *testing.T) {
	store := auth.NewSessionStore(time.Hour)
	r := setupSessionRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestSessionAuth_InvalidSID(t *testing.T) {
	store := auth.NewSessionStore(time.Hour)
	r := setupSessionRouter(store)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: "invalid-sid"})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestSessionAuth_Expired(t *testing.T) {
	store := auth.NewSessionStore(10 * time.Millisecond)
	sid := store.Create("10001", "alice", "")
	time.Sleep(50 * time.Millisecond)

	r := setupSessionRouter(store)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.AddCookie(&http.Cookie{Name: SessionCookieName, Value: sid})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired session, got %d", w.Code)
	}
}

func TestSessionStore_Delete(t *testing.T) {
	store := auth.NewSessionStore(time.Hour)
	sid := store.Create("10001", "alice", "")

	if store.Get(sid) == nil {
		t.Fatal("expected session to exist")
	}
	store.Delete(sid)
	if store.Get(sid) != nil {
		t.Fatal("expected session to be deleted")
	}
	// 幂等：再次 Delete 不 panic
	store.Delete(sid)
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOfStr(s, substr) >= 0))
}

func indexOfStr(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}