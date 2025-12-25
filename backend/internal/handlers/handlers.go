package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"payforwardnow/internal/database"
	"payforwardnow/internal/models"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"golang.org/x/crypto/bcrypt"
)

// Handler holds dependencies for HTTP handlers
type Handler struct {
	db database.DBClient
}

// NewHandler creates a new Handler
func NewHandler(db database.DBClient) *Handler {
	return &Handler{db: db}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check Neo4j connectivity
	session := h.db.ReadSession(ctx)
	defer session.Close(ctx)

	_, err := session.Run(ctx, "RETURN 1", nil)
	if err != nil {
		respondError(w, http.StatusServiceUnavailable, "DATABASE_ERROR", "Database connection failed")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "payforwardnow-api",
		"version":   "1.0.0",
	})
}

// GetUser handles GET /api/v1/users/{id}
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	ctx := r.Context()

	result, err := h.db.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (u:User {id: $id})
			OPTIONAL MATCH (u)-[:GAVE]->(given:Act)
			OPTIONAL MATCH (u)-[:RECEIVED]->(received:Act)
			OPTIONAL MATCH (u)-[:STARTED]->(chain:Chain)
			RETURN u, 
				   count(DISTINCT given) as actsGiven,
				   count(DISTINCT received) as actsReceived,
				   count(DISTINCT chain) as chainsStarted
		`
		result, err := tx.Run(ctx, query, map[string]interface{}{"id": userID})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			userNode, _ := record.Get("u")
			if userNode == nil {
				return nil, nil
			}

			node := userNode.(neo4j.Node)
			props := node.Props

			user := &models.User{
				ID:         props["id"].(string),
				Email:      props["email"].(string),
				Name:       props["name"].(string),
				IsVerified: props["isVerified"].(bool),
				CreatedAt:  props["createdAt"].(time.Time),
				UpdatedAt:  props["updatedAt"].(time.Time),
				Stats: models.UserStats{
					ActsGiven:     int(record.Values[1].(int64)),
					ActsReceived:  int(record.Values[2].(int64)),
					ChainsStarted: int(record.Values[3].(int64)),
				},
			}

			if avatar, ok := props["avatar"].(string); ok {
				user.Avatar = avatar
			}
			if bio, ok := props["bio"].(string); ok {
				user.Bio = bio
			}
			if location, ok := props["location"].(string); ok {
				user.Location = location
			}

			return user, nil
		}

		return nil, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch user")
		return
	}

	if result == nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// CreateUser handles POST /api/v1/users
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	ctx := r.Context()

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "HASH_ERROR", "Failed to hash password")
		return
	}

	now := time.Now().UTC()
	userID := uuid.New().String()

	result, err := h.db.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (u:User {
				id: $id,
				email: $email,
				passwordHash: $passwordHash,
				name: $name,
				isVerified: false,
				createdAt: $createdAt,
				updatedAt: $updatedAt
			})
			RETURN u
		`
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"id":           userID,
			"email":        req.Email,
			"passwordHash": string(hashedPassword),
			"name":         req.Name,
			"createdAt":    now,
			"updatedAt":    now,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return &models.User{
				ID:         userID,
				Email:      req.Email,
				Name:       req.Name,
				IsVerified: false,
				CreatedAt:  now,
				UpdatedAt:  now,
			}, nil
		}

		return nil, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create user")
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateUser handles PUT /api/v1/users/{id}
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	ctx := r.Context()

	result, err := h.db.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (u:User {id: $id})
			SET u.name = COALESCE($name, u.name),
				u.avatar = COALESCE($avatar, u.avatar),
				u.bio = COALESCE($bio, u.bio),
				u.location = COALESCE($location, u.location),
				u.updatedAt = $updatedAt
			RETURN u
		`
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"id":        userID,
			"name":      nilIfEmpty(req.Name),
			"avatar":    nilIfEmpty(req.Avatar),
			"bio":       nilIfEmpty(req.Bio),
			"location":  nilIfEmpty(req.Location),
			"updatedAt": time.Now().UTC(),
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update user")
		return
	}

	if result == false {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "User not found")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "User updated successfully"},
	})
}

// DeleteUser handles DELETE /api/v1/users/{id}
func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	ctx := r.Context()

	_, err := h.db.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (u:User {id: $id})
			DETACH DELETE u
		`
		_, err := tx.Run(ctx, query, map[string]interface{}{"id": userID})
		return nil, err
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete user")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "User deleted successfully"},
	})
}

// Register handles POST /api/v1/auth/register
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	ctx := r.Context()

	// Check if email exists
	exists, _ := h.db.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		result, err := tx.Run(ctx, "MATCH (u:User {email: $email}) RETURN u", map[string]interface{}{"email": req.Email})
		if err != nil {
			return false, err
		}
		return result.Next(ctx), nil
	})

	if exists.(bool) {
		respondError(w, http.StatusConflict, "EMAIL_EXISTS", "Email already registered")
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "HASH_ERROR", "Failed to hash password")
		return
	}

	now := time.Now().UTC()
	userID := uuid.New().String()

	_, err = h.db.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (u:User {
				id: $id,
				email: $email,
				passwordHash: $passwordHash,
				name: $name,
				isVerified: false,
				createdAt: $createdAt,
				updatedAt: $updatedAt
			})
			RETURN u
		`
		return tx.Run(ctx, query, map[string]interface{}{
			"id":           userID,
			"email":        req.Email,
			"passwordHash": string(hashedPassword),
			"name":         req.Name,
			"createdAt":    now,
			"updatedAt":    now,
		})
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create user")
		return
	}

	// Generate tokens (simplified - use proper JWT in production)
	tokens := models.AuthTokens{
		AccessToken:  uuid.New().String(),
		RefreshToken: uuid.New().String(),
		ExpiresIn:    3600,
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"user": models.User{
				ID:         userID,
				Email:      req.Email,
				Name:       req.Name,
				IsVerified: false,
				CreatedAt:  now,
				UpdatedAt:  now,
			},
			"tokens": tokens,
		},
	})
}

// Login handles POST /api/v1/auth/login
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	ctx := r.Context()

	result, err := h.db.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `MATCH (u:User {email: $email}) RETURN u`
		result, err := tx.Run(ctx, query, map[string]interface{}{"email": req.Email})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			userNode, _ := record.Get("u")
			node := userNode.(neo4j.Node)
			return node.Props, nil
		}
		return nil, nil
	})

	if err != nil || result == nil {
		respondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	props := result.(map[string]interface{})
	storedHash := props["passwordHash"].(string)

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(req.Password)); err != nil {
		respondError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	// Generate tokens
	tokens := models.AuthTokens{
		AccessToken:  uuid.New().String(),
		RefreshToken: uuid.New().String(),
		ExpiresIn:    3600,
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"user": models.User{
				ID:         props["id"].(string),
				Email:      props["email"].(string),
				Name:       props["name"].(string),
				IsVerified: props["isVerified"].(bool),
				CreatedAt:  props["createdAt"].(time.Time),
				UpdatedAt:  props["updatedAt"].(time.Time),
			},
			"tokens": tokens,
		},
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, invalidate the token
	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Logged out successfully"},
	})
}

// RefreshToken handles POST /api/v1/auth/refresh
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	// In a real implementation, validate refresh token and issue new tokens
	tokens := models.AuthTokens{
		AccessToken:  uuid.New().String(),
		RefreshToken: uuid.New().String(),
		ExpiresIn:    3600,
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    tokens,
	})
}

// GetActs handles GET /api/v1/acts
func (h *Handler) GetActs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := getPaginationParams(r)

	result, err := h.db.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		countQuery := `MATCH (a:Act) RETURN count(a) as total`
		countResult, err := tx.Run(ctx, countQuery, nil)
		if err != nil {
			return nil, err
		}

		var total int64
		if countResult.Next(ctx) {
			if val, ok := countResult.Record().Get("total"); ok {
				total = val.(int64)
			}
		}

		skip := (params.Page - 1) * params.PerPage
		query := `
			MATCH (a:Act)
			OPTIONAL MATCH (giver:User)-[:GAVE]->(a)
			OPTIONAL MATCH (a)-[:RECEIVED_BY]->(receiver:User)
			RETURN a, giver, receiver
			ORDER BY a.createdAt DESC
			SKIP $skip LIMIT $limit
		`
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"skip":  skip,
			"limit": params.PerPage,
		})
		if err != nil {
			return nil, err
		}

		var acts []models.Act
		for result.Next(ctx) {
			record := result.Record()
			actNode, _ := record.Get("a")
			node := actNode.(neo4j.Node)
			props := node.Props

			act := models.Act{
				ID:          props["id"].(string),
				Title:       props["title"].(string),
				Description: props["description"].(string),
				Type:        models.ActType(props["type"].(string)),
				Status:      models.ActStatus(props["status"].(string)),
				CreatedAt:   props["createdAt"].(time.Time),
				UpdatedAt:   props["updatedAt"].(time.Time),
			}

			if category, ok := props["category"].(string); ok {
				act.Category = category
			}
			if value, ok := props["value"].(float64); ok {
				act.Value = value
			}

			acts = append(acts, act)
		}

		return map[string]interface{}{
			"acts":  acts,
			"total": total,
		}, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch acts")
		return
	}

	data := result.(map[string]interface{})
	total := data["total"].(int64)
	totalPages := (int(total) + params.PerPage - 1) / params.PerPage

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    data["acts"],
		Meta: &models.APIMeta{
			Page:       params.Page,
			PerPage:    params.PerPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// CreateAct handles POST /api/v1/acts
func (h *Handler) CreateAct(w http.ResponseWriter, r *http.Request) {
	var req models.CreateActRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	ctx := r.Context()
	now := time.Now().UTC()
	actID := uuid.New().String()

	// Get user ID from context (should be set by auth middleware)
	giverID := r.Header.Get("X-User-ID")
	if giverID == "" {
		giverID = "anonymous"
	}

	result, err := h.db.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (a:Act {
				id: $id,
				title: $title,
				description: $description,
				type: $type,
				category: $category,
				value: $value,
				currency: $currency,
				status: 'pending',
				giverId: $giverId,
				receiverId: $receiverId,
				location: $location,
				isAnonymous: $isAnonymous,
				createdAt: $createdAt,
				updatedAt: $updatedAt
			})
			WITH a
			MATCH (giver:User {id: $giverId})
			CREATE (giver)-[:GAVE]->(a)
			RETURN a
		`
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"id":          actID,
			"title":       req.Title,
			"description": req.Description,
			"type":        string(req.Type),
			"category":    req.Category,
			"value":       req.Value,
			"currency":    req.Currency,
			"giverId":     giverID,
			"receiverId":  nilIfEmpty(req.ReceiverID),
			"location":    nilIfEmpty(req.Location),
			"isAnonymous": req.IsAnonymous,
			"createdAt":   now,
			"updatedAt":   now,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return &models.Act{
				ID:          actID,
				Title:       req.Title,
				Description: req.Description,
				Type:        req.Type,
				Category:    req.Category,
				Value:       req.Value,
				Status:      models.ActStatusPending,
				GiverID:     giverID,
				IsAnonymous: req.IsAnonymous,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		}
		return nil, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create act")
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetAct handles GET /api/v1/acts/{id}
func (h *Handler) GetAct(w http.ResponseWriter, r *http.Request) {
	actID := r.PathValue("id")
	ctx := r.Context()

	result, err := h.db.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (a:Act {id: $id})
			OPTIONAL MATCH (giver:User)-[:GAVE]->(a)
			OPTIONAL MATCH (a)-[:RECEIVED_BY]->(receiver:User)
			RETURN a, giver, receiver
		`
		result, err := tx.Run(ctx, query, map[string]interface{}{"id": actID})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			actNode, _ := record.Get("a")
			node := actNode.(neo4j.Node)
			props := node.Props

			act := &models.Act{
				ID:          props["id"].(string),
				Title:       props["title"].(string),
				Description: props["description"].(string),
				Type:        models.ActType(props["type"].(string)),
				Status:      models.ActStatus(props["status"].(string)),
				CreatedAt:   props["createdAt"].(time.Time),
				UpdatedAt:   props["updatedAt"].(time.Time),
			}

			return act, nil
		}

		return nil, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch act")
		return
	}

	if result == nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "Act not found")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// UpdateAct handles PUT /api/v1/acts/{id}
func (h *Handler) UpdateAct(w http.ResponseWriter, r *http.Request) {
	actID := r.PathValue("id")
	var req models.UpdateActRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	ctx := r.Context()

	_, err := h.db.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (a:Act {id: $id})
			SET a.title = COALESCE($title, a.title),
				a.description = COALESCE($description, a.description),
				a.status = COALESCE($status, a.status),
				a.updatedAt = $updatedAt
			RETURN a
		`
		return tx.Run(ctx, query, map[string]interface{}{
			"id":          actID,
			"title":       nilIfEmpty(req.Title),
			"description": nilIfEmpty(req.Description),
			"status":      nilIfEmpty(string(req.Status)),
			"updatedAt":   time.Now().UTC(),
		})
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update act")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Act updated successfully"},
	})
}

// DeleteAct handles DELETE /api/v1/acts/{id}
func (h *Handler) DeleteAct(w http.ResponseWriter, r *http.Request) {
	actID := r.PathValue("id")
	ctx := r.Context()

	_, err := h.db.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `MATCH (a:Act {id: $id}) DETACH DELETE a`
		return tx.Run(ctx, query, map[string]interface{}{"id": actID})
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete act")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Act deleted successfully"},
	})
}

// GetChain handles GET /api/v1/chains/{id}
func (h *Handler) GetChain(w http.ResponseWriter, r *http.Request) {
	chainID := r.PathValue("id")
	ctx := r.Context()

	result, err := h.db.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (c:Chain {id: $id})
			OPTIONAL MATCH (c)-[:CONTAINS]->(a:Act)
			OPTIONAL MATCH (starter:User)-[:STARTED]->(c)
			RETURN c, collect(a) as acts, starter
		`
		result, err := tx.Run(ctx, query, map[string]interface{}{"id": chainID})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			chainNode, _ := record.Get("c")
			if chainNode == nil {
				return nil, nil
			}

			node := chainNode.(neo4j.Node)
			props := node.Props

			chain := &models.Chain{
				ID:        props["id"].(string),
				Name:      props["name"].(string),
				CreatedAt: props["createdAt"].(time.Time),
				UpdatedAt: props["updatedAt"].(time.Time),
			}

			if desc, ok := props["description"].(string); ok {
				chain.Description = desc
			}

			return chain, nil
		}

		return nil, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch chain")
		return
	}

	if result == nil {
		respondError(w, http.StatusNotFound, "NOT_FOUND", "Chain not found")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetUserChains handles GET /api/v1/users/{id}/chains
func (h *Handler) GetUserChains(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	ctx := r.Context()

	result, err := h.db.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (u:User {id: $userId})-[:STARTED|PARTICIPATED_IN]->(c:Chain)
			RETURN DISTINCT c
			ORDER BY c.createdAt DESC
		`
		result, err := tx.Run(ctx, query, map[string]interface{}{"userId": userID})
		if err != nil {
			return nil, err
		}

		var chains []models.Chain
		for result.Next(ctx) {
			record := result.Record()
			chainNode, _ := record.Get("c")
			node := chainNode.(neo4j.Node)
			props := node.Props

			chain := models.Chain{
				ID:        props["id"].(string),
				Name:      props["name"].(string),
				CreatedAt: props["createdAt"].(time.Time),
				UpdatedAt: props["updatedAt"].(time.Time),
			}
			chains = append(chains, chain)
		}

		return chains, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch chains")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetGlobalStats handles GET /api/v1/stats/global
func (h *Handler) GetGlobalStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.db.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (a:Act)
			WITH count(a) as totalActs, sum(COALESCE(a.value, 0)) as totalValue
			MATCH (u:User)
			WITH totalActs, totalValue, count(u) as totalUsers
			MATCH (c:Chain)
			RETURN totalActs, totalValue, totalUsers, count(c) as totalChains
		`
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			return &models.GlobalStats{
				TotalActs:      getInt64(record, "totalActs"),
				TotalValue:     getFloat64(record, "totalValue"),
				TotalUsers:     getInt64(record, "totalUsers"),
				TotalChains:    getInt64(record, "totalChains"),
				CountriesReach: 180, // Placeholder
			}, nil
		}

		return &models.GlobalStats{}, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch stats")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetUserStats handles GET /api/v1/stats/user/{id}
func (h *Handler) GetUserStats(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")
	ctx := r.Context()

	result, err := h.db.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (u:User {id: $userId})
			OPTIONAL MATCH (u)-[:GAVE]->(given:Act)
			OPTIONAL MATCH (u)-[:RECEIVED]->(received:Act)
			OPTIONAL MATCH (u)-[:STARTED]->(chain:Chain)
			RETURN 
				count(DISTINCT given) as actsGiven,
				count(DISTINCT received) as actsReceived,
				count(DISTINCT chain) as chainsStarted,
				sum(COALESCE(given.value, 0)) as totalImpact
		`
		result, err := tx.Run(ctx, query, map[string]interface{}{"userId": userID})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			record := result.Record()
			return &models.UserStats{
				ActsGiven:     int(getInt64(record, "actsGiven")),
				ActsReceived:  int(getInt64(record, "actsReceived")),
				ChainsStarted: int(getInt64(record, "chainsStarted")),
				TotalImpact:   getFloat64(record, "totalImpact"),
			}, nil
		}

		return &models.UserStats{}, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch stats")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// GetTestimonials handles GET /api/v1/testimonials
func (h *Handler) GetTestimonials(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.db.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (t:Testimonial {isApproved: true})
			OPTIONAL MATCH (u:User)-[:WROTE]->(t)
			RETURN t, u
			ORDER BY t.createdAt DESC
			LIMIT 20
		`
		result, err := tx.Run(ctx, query, nil)
		if err != nil {
			return nil, err
		}

		var testimonials []models.Testimonial
		for result.Next(ctx) {
			record := result.Record()
			testNode, _ := record.Get("t")
			node := testNode.(neo4j.Node)
			props := node.Props

			testimonial := models.Testimonial{
				ID:         props["id"].(string),
				Story:      props["story"].(string),
				Impact:     props["impact"].(string),
				IsApproved: props["isApproved"].(bool),
				CreatedAt:  props["createdAt"].(time.Time),
			}

			if userNode, ok := record.Get("u"); ok && userNode != nil {
				uNode := userNode.(neo4j.Node)
				uProps := uNode.Props
				testimonial.User = &models.User{
					ID:   uProps["id"].(string),
					Name: uProps["name"].(string),
				}
				if location, ok := uProps["location"].(string); ok {
					testimonial.User.Location = location
				}
			}

			testimonials = append(testimonials, testimonial)
		}

		return testimonials, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch testimonials")
		return
	}

	respondJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// CreateTestimonial handles POST /api/v1/testimonials
func (h *Handler) CreateTestimonial(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTestimonialRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	ctx := r.Context()
	now := time.Now().UTC()
	testID := uuid.New().String()
	userID := r.Header.Get("X-User-ID")

	result, err := h.db.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			CREATE (t:Testimonial {
				id: $id,
				userId: $userId,
				story: $story,
				impact: $impact,
				isApproved: false,
				isFeatured: false,
				createdAt: $createdAt
			})
			WITH t
			MATCH (u:User {id: $userId})
			CREATE (u)-[:WROTE]->(t)
			RETURN t
		`
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"id":        testID,
			"userId":    userID,
			"story":     req.Story,
			"impact":    req.Impact,
			"createdAt": now,
		})
		if err != nil {
			return nil, err
		}

		if result.Next(ctx) {
			return &models.Testimonial{
				ID:         testID,
				UserID:     userID,
				Story:      req.Story,
				Impact:     req.Impact,
				IsApproved: false,
				IsFeatured: false,
				CreatedAt:  now,
			}, nil
		}
		return nil, nil
	})

	if err != nil {
		respondError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create testimonial")
		return
	}

	respondJSON(w, http.StatusCreated, models.APIResponse{
		Success: true,
		Data:    result,
	})
}

// Helper functions
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, code, message string) {
	respondJSON(w, status, models.APIResponse{
		Success: false,
		Error: &models.APIError{
			Code:    code,
			Message: message,
		},
	})
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func getPaginationParams(r *http.Request) models.PaginationParams {
	params := models.DefaultPagination()

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}

	if perPage := r.URL.Query().Get("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil && pp > 0 && pp <= 100 {
			params.PerPage = pp
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		params.SortBy = sortBy
	}

	if order := r.URL.Query().Get("order"); order == "asc" || order == "desc" {
		params.Order = order
	}

	return params
}

func getInt64(record *neo4j.Record, key string) int64 {
	if val, ok := record.Get(key); ok && val != nil {
		return val.(int64)
	}
	return 0
}

func getFloat64(record *neo4j.Record, key string) float64 {
	if val, ok := record.Get(key); ok && val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case int64:
			return float64(v)
		}
	}
	return 0
}
