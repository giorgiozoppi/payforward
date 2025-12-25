package models

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Avatar       string    `json:"avatar,omitempty"`
	Bio          string    `json:"bio,omitempty"`
	Location     string    `json:"location,omitempty"`
	IsVerified   bool      `json:"isVerified"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
	Stats        UserStats `json:"stats,omitempty"`
}

// UserStats holds user statistics
type UserStats struct {
	ActsGiven     int     `json:"actsGiven"`
	ActsReceived  int     `json:"actsReceived"`
	ChainsStarted int     `json:"chainsStarted"`
	TotalImpact   float64 `json:"totalImpact"`
	KindnessScore float64 `json:"kindnessScore"`
	ActiveStreak  int     `json:"activeStreak"`
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

// UpdateUserRequest represents a request to update a user
type UpdateUserRequest struct {
	Name     string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Avatar   string `json:"avatar,omitempty"`
	Bio      string `json:"bio,omitempty" validate:"omitempty,max=500"`
	Location string `json:"location,omitempty" validate:"omitempty,max=100"`
}

// Act represents an act of kindness
type Act struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Type        ActType    `json:"type"`
	Category    string     `json:"category"`
	Value       float64    `json:"value,omitempty"`
	Currency    string     `json:"currency,omitempty"`
	Status      ActStatus  `json:"status"`
	GiverID     string     `json:"giverId"`
	ReceiverID  string     `json:"receiverId,omitempty"`
	ChainID     string     `json:"chainId,omitempty"`
	Location    string     `json:"location,omitempty"`
	IsAnonymous bool       `json:"isAnonymous"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
	Giver       *User      `json:"giver,omitempty"`
	Receiver    *User      `json:"receiver,omitempty"`
}

// ActType represents the type of act
type ActType string

const (
	ActTypeMonetary  ActType = "monetary"
	ActTypeService   ActType = "service"
	ActTypeGoods     ActType = "goods"
	ActTypeMentoring ActType = "mentoring"
	ActTypeOther     ActType = "other"
)

// ActStatus represents the status of an act
type ActStatus string

const (
	ActStatusPending   ActStatus = "pending"
	ActStatusAccepted  ActStatus = "accepted"
	ActStatusCompleted ActStatus = "completed"
	ActStatusCancelled ActStatus = "cancelled"
)

// CreateActRequest represents a request to create an act
type CreateActRequest struct {
	Title       string  `json:"title" validate:"required,min=5,max=200"`
	Description string  `json:"description" validate:"required,min=10,max=2000"`
	Type        ActType `json:"type" validate:"required"`
	Category    string  `json:"category" validate:"required"`
	Value       float64 `json:"value,omitempty"`
	Currency    string  `json:"currency,omitempty"`
	ReceiverID  string  `json:"receiverId,omitempty"`
	Location    string  `json:"location,omitempty"`
	IsAnonymous bool    `json:"isAnonymous"`
}

// UpdateActRequest represents a request to update an act
type UpdateActRequest struct {
	Title       string    `json:"title,omitempty" validate:"omitempty,min=5,max=200"`
	Description string    `json:"description,omitempty" validate:"omitempty,min=10,max=2000"`
	Status      ActStatus `json:"status,omitempty"`
	ReceiverID  string    `json:"receiverId,omitempty"`
}

// Chain represents a chain of kindness
type Chain struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	StarterID   string    `json:"starterId"`
	ActsCount   int       `json:"actsCount"`
	TotalValue  float64   `json:"totalValue"`
	Reach       int       `json:"reach"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	Acts        []Act     `json:"acts,omitempty"`
	Starter     *User     `json:"starter,omitempty"`
}

// Testimonial represents a user testimonial
type Testimonial struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	Story      string    `json:"story"`
	Impact     string    `json:"impact"`
	IsApproved bool      `json:"isApproved"`
	IsFeatured bool      `json:"isFeatured"`
	CreatedAt  time.Time `json:"createdAt"`
	User       *User     `json:"user,omitempty"`
}

// CreateTestimonialRequest represents a request to create a testimonial
type CreateTestimonialRequest struct {
	Story  string `json:"story" validate:"required,min=50,max=2000"`
	Impact string `json:"impact" validate:"required,min=10,max=200"`
}

// GlobalStats represents global platform statistics
type GlobalStats struct {
	TotalActs       int64   `json:"totalActs"`
	TotalUsers      int64   `json:"totalUsers"`
	TotalChains     int64   `json:"totalChains"`
	TotalValue      float64 `json:"totalValue"`
	CountriesReach  int     `json:"countriesReach"`
	ActiveThisMonth int64   `json:"activeThisMonth"`
}

// AuthTokens represents authentication tokens
type AuthTokens struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *APIMeta    `json:"meta,omitempty"`
}

// APIError represents an API error
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// APIMeta represents API metadata
type APIMeta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"perPage,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"totalPages,omitempty"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page    int    `json:"page"`
	PerPage int    `json:"perPage"`
	SortBy  string `json:"sortBy"`
	Order   string `json:"order"`
}

// DefaultPagination returns default pagination params
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:    1,
		PerPage: 20,
		SortBy:  "createdAt",
		Order:   "desc",
	}
}
