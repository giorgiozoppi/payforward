package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"payforwardnow/internal/models"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// MockDBClient is a mock implementation of the DBClient interface for testing
type MockDBClient struct {
	ExecuteReadFunc  func(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error)
	ExecuteWriteFunc func(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error)
	ReadSessionFunc  func(ctx context.Context) neo4j.SessionWithContext
}

func (m *MockDBClient) ExecuteRead(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	if m.ExecuteReadFunc != nil {
		return m.ExecuteReadFunc(ctx, work)
	}
	return work(nil)
}

func (m *MockDBClient) ExecuteWrite(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
	if m.ExecuteWriteFunc != nil {
		return m.ExecuteWriteFunc(ctx, work)
	}
	return work(nil)
}

func (m *MockDBClient) ReadSession(ctx context.Context) neo4j.SessionWithContext {
	if m.ReadSessionFunc != nil {
		return m.ReadSessionFunc(ctx)
	}
	return nil
}

func (m *MockDBClient) Close() error {
	return nil
}

func TestRegister_EmailAlreadyExists(t *testing.T) {
	mockDB := &MockDBClient{
		ExecuteReadFunc: func(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
			return true, nil
		},
	}
	handler := NewHandler(mockDB)

	requestBody := models.RegisterRequest{
		Email:    "existing@example.com",
		Password: "password123",
		Name:     "Test User",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, w.Code)
	}

	var response models.APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Success {
		t.Error("expected success to be false")
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	mockDB := &MockDBClient{}
	handler := NewHandler(mockDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	mockDB := &MockDBClient{
		ExecuteReadFunc: func(ctx context.Context, work func(tx neo4j.ManagedTransaction) (interface{}, error)) (interface{}, error) {
			return nil, nil
		},
	}
	handler := NewHandler(mockDB)

	requestBody := models.LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var response models.APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Success {
		t.Error("expected success to be false")
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	mockDB := &MockDBClient{}
	handler := NewHandler(mockDB)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestLogout(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("expected success to be true")
	}
}

func TestRefreshToken(t *testing.T) {
	handler := NewHandler(nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	w := httptest.NewRecorder()

	handler.RefreshToken(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response models.APIResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("expected success to be true")
	}

	tokens, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("expected data to be tokens object")
	}

	if tokens["accessToken"] == nil || tokens["refreshToken"] == nil {
		t.Error("expected tokens to be present")
	}
}

func TestNilIfEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{
			name:     "empty string returns nil",
			input:    "",
			expected: nil,
		},
		{
			name:     "non-empty string returns string",
			input:    "test",
			expected: "test",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nilIfEmpty(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetPaginationParams(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected models.PaginationParams
	}{
		{
			name: "default pagination",
			url:  "/api/v1/acts",
			expected: models.PaginationParams{
				Page:    1,
				PerPage: 20,
				SortBy:  "createdAt",
				Order:   "desc",
			},
		},
		{
			name: "custom pagination",
			url:  "/api/v1/acts?page=2&per_page=10&sort_by=title&order=asc",
			expected: models.PaginationParams{
				Page:    2,
				PerPage: 10,
				SortBy:  "title",
				Order:   "asc",
			},
		},
		{
			name: "invalid page defaults to 1",
			url:  "/api/v1/acts?page=-1&per_page=10",
			expected: models.PaginationParams{
				Page:    1,
				PerPage: 10,
				SortBy:  "createdAt",
				Order:   "desc",
			},
		},
		{
			name: "per_page over 100 defaults to 20",
			url:  "/api/v1/acts?per_page=200",
			expected: models.PaginationParams{
				Page:    1,
				PerPage: 20,
				SortBy:  "createdAt",
				Order:   "desc",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			result := getPaginationParams(req)

			if result.Page != tt.expected.Page {
				t.Errorf("expected page %d, got %d", tt.expected.Page, result.Page)
			}
			if result.PerPage != tt.expected.PerPage {
				t.Errorf("expected perPage %d, got %d", tt.expected.PerPage, result.PerPage)
			}
			if result.SortBy != tt.expected.SortBy {
				t.Errorf("expected sortBy %s, got %s", tt.expected.SortBy, result.SortBy)
			}
			if result.Order != tt.expected.Order {
				t.Errorf("expected order %s, got %s", tt.expected.Order, result.Order)
			}
		})
	}
}
