package models

import (
	"encoding/json"
	"testing"
)

func TestDefaultPagination(t *testing.T) {
	params := DefaultPagination()

	if params.Page != 1 {
		t.Errorf("expected page 1, got %d", params.Page)
	}
	if params.PerPage != 20 {
		t.Errorf("expected perPage 20, got %d", params.PerPage)
	}
	if params.SortBy != "createdAt" {
		t.Errorf("expected sortBy 'createdAt', got %s", params.SortBy)
	}
	if params.Order != "desc" {
		t.Errorf("expected order 'desc', got %s", params.Order)
	}
}

func TestActType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		actType  ActType
		expected string
	}{
		{"monetary", ActTypeMonetary, "monetary"},
		{"service", ActTypeService, "service"},
		{"goods", ActTypeGoods, "goods"},
		{"mentoring", ActTypeMentoring, "mentoring"},
		{"other", ActTypeOther, "other"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.actType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.actType))
			}
		})
	}
}

func TestActStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		status   ActStatus
		expected string
	}{
		{"pending", ActStatusPending, "pending"},
		{"accepted", ActStatusAccepted, "accepted"},
		{"completed", ActStatusCompleted, "completed"},
		{"cancelled", ActStatusCancelled, "cancelled"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.status))
			}
		})
	}
}

func TestUserJSON(t *testing.T) {
	user := User{
		ID:         "user-123",
		Email:      "test@example.com",
		Name:       "Test User",
		IsVerified: true,
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	var decoded User
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal user: %v", err)
	}

	if decoded.ID != user.ID {
		t.Errorf("expected ID %s, got %s", user.ID, decoded.ID)
	}
	if decoded.Email != user.Email {
		t.Errorf("expected email %s, got %s", user.Email, decoded.Email)
	}
	if decoded.Name != user.Name {
		t.Errorf("expected name %s, got %s", user.Name, decoded.Name)
	}
	if decoded.IsVerified != user.IsVerified {
		t.Errorf("expected isVerified %v, got %v", user.IsVerified, decoded.IsVerified)
	}
}

func TestUserJSON_PasswordHashOmitted(t *testing.T) {
	user := User{
		ID:           "user-123",
		Email:        "test@example.com",
		PasswordHash: "secret-hash",
		Name:         "Test User",
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("failed to marshal user: %v", err)
	}

	jsonStr := string(data)
	if jsonContains(jsonStr, "passwordHash") || jsonContains(jsonStr, "secret-hash") {
		t.Error("passwordHash should not be included in JSON")
	}
}

func TestActJSON(t *testing.T) {
	act := Act{
		ID:          "act-123",
		Title:       "Test Act",
		Description: "Test Description",
		Type:        ActTypeMonetary,
		Status:      ActStatusPending,
	}

	data, err := json.Marshal(act)
	if err != nil {
		t.Fatalf("failed to marshal act: %v", err)
	}

	var decoded Act
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal act: %v", err)
	}

	if decoded.ID != act.ID {
		t.Errorf("expected ID %s, got %s", act.ID, decoded.ID)
	}
	if decoded.Title != act.Title {
		t.Errorf("expected title %s, got %s", act.Title, decoded.Title)
	}
	if decoded.Type != act.Type {
		t.Errorf("expected type %s, got %s", act.Type, decoded.Type)
	}
	if decoded.Status != act.Status {
		t.Errorf("expected status %s, got %s", act.Status, decoded.Status)
	}
}

func TestAPIResponse(t *testing.T) {
	response := APIResponse{
		Success: true,
		Data:    map[string]string{"message": "success"},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded APIResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.Success != response.Success {
		t.Errorf("expected success %v, got %v", response.Success, decoded.Success)
	}
}

func TestAPIResponse_WithError(t *testing.T) {
	response := APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "TEST_ERROR",
			Message: "Test error message",
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded APIResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.Success != false {
		t.Error("expected success to be false")
	}
	if decoded.Error == nil {
		t.Fatal("expected error to be present")
	}

	// Re-marshal and unmarshal to get the error as map
	data2, _ := json.Marshal(decoded.Error)
	var errorData map[string]interface{}
	json.Unmarshal(data2, &errorData)
	if errorData["code"] != "TEST_ERROR" {
		t.Errorf("expected error code TEST_ERROR, got %v", errorData["code"])
	}
}

func TestAPIResponse_WithMeta(t *testing.T) {
	response := APIResponse{
		Success: true,
		Data:    []string{"item1", "item2"},
		Meta: &APIMeta{
			Page:       1,
			PerPage:    20,
			Total:      100,
			TotalPages: 5,
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var decoded APIResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if decoded.Meta == nil {
		t.Fatal("expected meta to be present")
	}

	// Re-marshal and unmarshal to get the meta as map
	data2, _ := json.Marshal(decoded.Meta)
	var metaData map[string]interface{}
	json.Unmarshal(data2, &metaData)
	if metaData["page"].(float64) != 1 {
		t.Errorf("expected page 1, got %v", metaData["page"])
	}
	if metaData["totalPages"].(float64) != 5 {
		t.Errorf("expected totalPages 5, got %v", metaData["totalPages"])
	}
}

func TestCreateUserRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		request CreateUserRequest
		valid   bool
	}{
		{
			name: "valid request",
			request: CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "Test User",
			},
			valid: true,
		},
		{
			name: "empty email",
			request: CreateUserRequest{
				Email:    "",
				Password: "password123",
				Name:     "Test User",
			},
			valid: false,
		},
		{
			name: "short password",
			request: CreateUserRequest{
				Email:    "test@example.com",
				Password: "short",
				Name:     "Test User",
			},
			valid: false,
		},
		{
			name: "short name",
			request: CreateUserRequest{
				Email:    "test@example.com",
				Password: "password123",
				Name:     "A",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			var decoded CreateUserRequest
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("failed to unmarshal request: %v", err)
			}

			if tt.valid {
				if decoded.Email == "" || decoded.Password == "" || decoded.Name == "" {
					t.Error("expected all fields to be populated")
				}
			}
		})
	}
}

func TestCreateActRequest(t *testing.T) {
	request := CreateActRequest{
		Title:       "Test Act",
		Description: "Test Description that is long enough",
		Type:        ActTypeService,
		Category:    "helping",
		IsAnonymous: true,
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	var decoded CreateActRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal request: %v", err)
	}

	if decoded.Title != request.Title {
		t.Errorf("expected title %s, got %s", request.Title, decoded.Title)
	}
	if decoded.IsAnonymous != request.IsAnonymous {
		t.Errorf("expected isAnonymous %v, got %v", request.IsAnonymous, decoded.IsAnonymous)
	}
}

func TestAuthTokens(t *testing.T) {
	tokens := AuthTokens{
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-456",
		ExpiresIn:    3600,
	}

	data, err := json.Marshal(tokens)
	if err != nil {
		t.Fatalf("failed to marshal tokens: %v", err)
	}

	var decoded AuthTokens
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal tokens: %v", err)
	}

	if decoded.AccessToken != tokens.AccessToken {
		t.Errorf("expected accessToken %s, got %s", tokens.AccessToken, decoded.AccessToken)
	}
	if decoded.RefreshToken != tokens.RefreshToken {
		t.Errorf("expected refreshToken %s, got %s", tokens.RefreshToken, decoded.RefreshToken)
	}
	if decoded.ExpiresIn != tokens.ExpiresIn {
		t.Errorf("expected expiresIn %d, got %d", tokens.ExpiresIn, decoded.ExpiresIn)
	}
}

func TestGlobalStats(t *testing.T) {
	stats := GlobalStats{
		TotalActs:      1000,
		TotalUsers:     500,
		TotalChains:    50,
		TotalValue:     25000.50,
		CountriesReach: 42,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("failed to marshal stats: %v", err)
	}

	var decoded GlobalStats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal stats: %v", err)
	}

	if decoded.TotalActs != stats.TotalActs {
		t.Errorf("expected totalActs %d, got %d", stats.TotalActs, decoded.TotalActs)
	}
	if decoded.TotalValue != stats.TotalValue {
		t.Errorf("expected totalValue %f, got %f", stats.TotalValue, decoded.TotalValue)
	}
}

func TestUserStats(t *testing.T) {
	stats := UserStats{
		ActsGiven:     10,
		ActsReceived:  5,
		ChainsStarted: 2,
		TotalImpact:   1500.75,
		KindnessScore: 95.5,
		ActiveStreak:  7,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("failed to marshal stats: %v", err)
	}

	var decoded UserStats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal stats: %v", err)
	}

	if decoded.ActsGiven != stats.ActsGiven {
		t.Errorf("expected actsGiven %d, got %d", stats.ActsGiven, decoded.ActsGiven)
	}
	if decoded.TotalImpact != stats.TotalImpact {
		t.Errorf("expected totalImpact %f, got %f", stats.TotalImpact, decoded.TotalImpact)
	}
}

func jsonContains(jsonStr, substr string) bool {
	return len(jsonStr) > 0 && len(substr) > 0 &&
		(jsonStr == substr ||
			len(jsonStr) >= len(substr) &&
				(jsonStr[:len(substr)] == substr ||
					jsonStr[len(jsonStr)-len(substr):] == substr ||
					containsSubstring(jsonStr, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
