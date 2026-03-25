package routers

import (
	"calendar/internal/auth/jwt"
	"calendar/internal/domain/user"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func init() { gin.SetMode(gin.TestMode) }

// stubAuth returns a valid payload instantly — isolates middleware overhead.
type stubAuth struct{}

func (s *stubAuth) Login(context.Context, string, string) (*jwt.AuthTokens, error) { return nil, nil }
func (s *stubAuth) Register(context.Context, string, string, string, string) (*user.User, error) {
	return nil, nil
}
func (s *stubAuth) RefreshTokens(string) (*jwt.AuthTokens, error) { return nil, nil }
func (s *stubAuth) ValidateTokens(string) (*jwt.Payload, error) {
	return &jwt.Payload{UserID: uuid.New().String()}, nil
}

func benchAuthRouter() *gin.Engine {
	r := gin.New()
	r.Use(AuthMiddleware(&stubAuth{}))
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})
	return r
}

// BenchmarkAuthMiddleware measures the full middleware + tiny handler path.
func BenchmarkAuthMiddleware(b *testing.B) {
	router := benchAuthRouter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		router.ServeHTTP(w, req)
	}
}

// BenchmarkAuthMiddleware_Parallel measures concurrency safety.
func BenchmarkAuthMiddleware_Parallel(b *testing.B) {
	router := benchAuthRouter()

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer valid-token")
			router.ServeHTTP(w, req)
		}
	})
}

// BenchmarkAuthMiddleware_Reject measures the fast-reject path (no header).
func BenchmarkAuthMiddleware_Reject(b *testing.B) {
	router := benchAuthRouter()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
		// No Authorization header — should reject fast.
		router.ServeHTTP(w, req)
	}
}
