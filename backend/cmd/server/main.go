package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"payforwardnow/internal/auth"
	"payforwardnow/internal/database"
	"payforwardnow/internal/handlers"
	"payforwardnow/internal/middleware"
)

func main() {
	// Load configuration
	config := LoadConfig()

	// Initialize Neo4j connection
	neo4jClient, err := database.NewNeo4jClient(config.Neo4jURI, config.Neo4jUser, config.Neo4jPassword)
	if err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	defer neo4jClient.Close()

	// Initialize Keycloak authentication (if configured)
	var keycloakAuth *auth.KeycloakAuth
	if config.KeycloakURL != "" && config.KeycloakRealm != "" {
		keycloakAuth = auth.NewKeycloakAuth(
			config.KeycloakURL,
			config.KeycloakRealm,
			config.KeycloakClientID,
			config.KeycloakClientSecret,
		)
		_ = middleware.NewKeycloakAuthMiddleware(keycloakAuth)
		log.Printf("Keycloak authentication enabled for realm: %s", config.KeycloakRealm)
	} else {
		log.Println("Keycloak authentication disabled, using JWT tokens")
	}

	// Initialize handlers
	h := handlers.NewHandler(neo4jClient)

	// Setup router
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/health", h.HealthCheck)
	mux.HandleFunc("GET /api/v1/users/{id}", h.GetUser)
	mux.HandleFunc("POST /api/v1/users", h.CreateUser)
	mux.HandleFunc("PUT /api/v1/users/{id}", h.UpdateUser)
	mux.HandleFunc("DELETE /api/v1/users/{id}", h.DeleteUser)

	// Auth routes
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.RefreshToken)

	// Pay it forward routes
	mux.HandleFunc("GET /api/v1/acts", h.GetActs)
	mux.HandleFunc("POST /api/v1/acts", h.CreateAct)
	mux.HandleFunc("GET /api/v1/acts/{id}", h.GetAct)
	mux.HandleFunc("PUT /api/v1/acts/{id}", h.UpdateAct)
	mux.HandleFunc("DELETE /api/v1/acts/{id}", h.DeleteAct)

	// Chain routes
	mux.HandleFunc("GET /api/v1/chains/{id}", h.GetChain)
	mux.HandleFunc("GET /api/v1/users/{id}/chains", h.GetUserChains)

	// Stats routes
	mux.HandleFunc("GET /api/v1/stats/global", h.GetGlobalStats)
	mux.HandleFunc("GET /api/v1/stats/user/{id}", h.GetUserStats)

	// Testimonials routes
	mux.HandleFunc("GET /api/v1/testimonials", h.GetTestimonials)
	mux.HandleFunc("POST /api/v1/testimonials", h.CreateTestimonial)

	// Apply middleware stack
	handler := middleware.Chain(
		mux,
		middleware.Logger,
		middleware.CORS(config.AllowedOrigins),
		middleware.RateLimit(config.RateLimitPerMin),
		middleware.Recovery,
		middleware.SecurityHeaders,
		middleware.RequestID,
	)

	// Create server
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Printf("Server starting on port %s", config.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

// Config holds the application configuration
type Config struct {
	Port                 string
	Neo4jURI             string
	Neo4jUser            string
	Neo4jPassword        string
	JWTSecret            string
	Environment          string
	KeycloakURL          string
	KeycloakRealm        string
	KeycloakClientID     string
	KeycloakClientSecret string
	AllowedOrigins       []string
	RateLimitPerMin      int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	allowedOrigins := []string{"*"}
	if origins := getEnv("ALLOWED_ORIGINS", ""); origins != "" {
		allowedOrigins = strings.Split(origins, ",")
	}

	rateLimitPerMin := 100
	if limit := getEnv("RATE_LIMIT_PER_MIN", ""); limit != "" {
		if val, err := strconv.Atoi(limit); err == nil {
			rateLimitPerMin = val
		}
	}

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		Neo4jURI:             getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:            getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword:        getEnv("NEO4J_PASSWORD", "password"),
		JWTSecret:            getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		Environment:          getEnv("ENVIRONMENT", "development"),
		KeycloakURL:          getEnv("KEYCLOAK_URL", ""),
		KeycloakRealm:        getEnv("KEYCLOAK_REALM", ""),
		KeycloakClientID:     getEnv("KEYCLOAK_CLIENT_ID", ""),
		KeycloakClientSecret: getEnv("KEYCLOAK_CLIENT_SECRET", ""),
		AllowedOrigins:       allowedOrigins,
		RateLimitPerMin:      rateLimitPerMin,
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
