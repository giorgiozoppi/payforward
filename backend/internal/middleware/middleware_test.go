package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestChain(t *testing.T) {
	called := []string{}

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = append(called, "m1-before")
			next.ServeHTTP(w, r)
			called = append(called, "m1-after")
		})
	}

	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = append(called, "m2-before")
			next.ServeHTTP(w, r)
			called = append(called, "m2-after")
		})
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = append(called, "handler")
	})

	chained := Chain(handler, m1, m2)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	chained.ServeHTTP(w, req)

	expected := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
	if len(called) != len(expected) {
		t.Errorf("expected %d calls, got %d", len(expected), len(called))
	}

	for i, v := range expected {
		if called[i] != v {
			t.Errorf("expected call %d to be %s, got %s", i, v, called[i])
		}
	}
}

func TestLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := Logger(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCORS(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		requestOrigin  string
		expectOrigin   string
	}{
		{
			name:           "wildcard allows all",
			allowedOrigins: []string{"*"},
			requestOrigin:  "https://example.com",
			expectOrigin:   "https://example.com",
		},
		{
			name:           "specific origin allowed",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://example.com",
			expectOrigin:   "https://example.com",
		},
		{
			name:           "origin not allowed",
			allowedOrigins: []string{"https://example.com"},
			requestOrigin:  "https://evil.com",
			expectOrigin:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			middleware := CORS(tt.allowedOrigins)(handler)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Origin", tt.requestOrigin)
			w := httptest.NewRecorder()

			middleware.ServeHTTP(w, req)

			origin := w.Header().Get("Access-Control-Allow-Origin")
			if origin != tt.expectOrigin {
				t.Errorf("expected origin %s, got %s", tt.expectOrigin, origin)
			}
		})
	}
}

func TestCORS_PreflightRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS request")
	})

	middleware := CORS([]string{"*"})(handler)

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("expected Access-Control-Allow-Methods header to be set")
	}
}

func TestRateLimit(t *testing.T) {
	requestsPerMinute := 2
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := RateLimit(requestsPerMinute)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"

	for i := 0; i < requestsPerMinute; i++ {
		w := httptest.NewRecorder()
		middleware.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("request %d: expected status %d, got %d", i+1, http.StatusOK, w.Code)
		}
	}

	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d after rate limit, got %d", http.StatusTooManyRequests, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Rate limit exceeded") {
		t.Error("expected rate limit error message")
	}
}

func TestRecovery(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	middleware := Recovery(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	if !strings.Contains(w.Body.String(), "Internal server error") {
		t.Error("expected internal server error message")
	}
}

func TestGenerateToken(t *testing.T) {
	secret := "test-secret"
	userID := "user-123"
	email := "test@example.com"
	duration := time.Hour

	token, err := GenerateToken(secret, userID, email, duration)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestJWTAuth_Valid(t *testing.T) {
	secret := "test-secret"
	userID := "user-123"
	email := "test@example.com"

	token, err := GenerateToken(secret, userID, email, time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxUserID := r.Context().Value(UserIDKey)
		ctxEmail := r.Context().Value(EmailKey)

		if ctxUserID != userID {
			t.Errorf("expected userID %s, got %v", userID, ctxUserID)
		}
		if ctxEmail != email {
			t.Errorf("expected email %s, got %v", email, ctxEmail)
		}

		w.WriteHeader(http.StatusOK)
	})

	middleware := JWTAuth(secret)(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestJWTAuth_MissingToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called without token")
	})

	middleware := JWTAuth("test-secret")(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid token format")
	})

	middleware := JWTAuth("test-secret")(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "invalid-format")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid token")
	})

	middleware := JWTAuth("test-secret")(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestSecurityHeaders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	middleware := SecurityHeaders(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	headers := []string{
		"X-Content-Type-Options",
		"X-Frame-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Content-Security-Policy",
	}

	for _, header := range headers {
		if w.Header().Get(header) == "" {
			t.Errorf("expected %s header to be set", header)
		}
	}
}

func TestRequestID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(ContextKey("requestID"))
		if requestID == nil || requestID == "" {
			t.Error("expected request ID in context")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

func TestRequestID_ExistingID(t *testing.T) {
	existingID := "existing-request-id"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(ContextKey("requestID"))
		if requestID != existingID {
			t.Errorf("expected request ID %s, got %v", existingID, requestID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestID(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != existingID {
		t.Errorf("expected X-Request-ID %s, got %s", existingID, w.Header().Get("X-Request-ID"))
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	limiter := NewRateLimiter(10)

	limiter.allow("192.168.1.1")
	limiter.allow("192.168.1.2")

	if len(limiter.visitors) != 2 {
		t.Errorf("expected 2 visitors, got %d", len(limiter.visitors))
	}

	for ip := range limiter.visitors {
		limiter.visitors[ip].lastReset = time.Now().Add(-11 * time.Minute)
	}

	limiter.cleanup()

	if len(limiter.visitors) != 0 {
		t.Errorf("expected 0 visitors after cleanup, got %d", len(limiter.visitors))
	}
}

func TestResponseWrapper(t *testing.T) {
	w := httptest.NewRecorder()
	wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

	wrapper.WriteHeader(http.StatusNotFound)

	if wrapper.statusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, wrapper.statusCode)
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("expected underlying status code %d, got %d", http.StatusNotFound, w.Code)
	}
}
