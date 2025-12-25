package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"payforwardnow/internal/auth"
)

type KeycloakAuthMiddleware struct {
	keycloak *auth.KeycloakAuth
}

func NewKeycloakAuthMiddleware(keycloak *auth.KeycloakAuth) *KeycloakAuthMiddleware {
	return &KeycloakAuthMiddleware{
		keycloak: keycloak,
	}
}

// KeycloakAuth is a middleware that validates Keycloak JWT tokens
func (k *KeycloakAuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := k.keycloak.ExtractBearerToken(r)
		if err != nil {
			respondJSONError(w, http.StatusUnauthorized, "Missing or invalid authorization header")
			return
		}

		claims, err := k.keycloak.ValidateToken(tokenString)
		if err != nil {
			log.Printf("Token validation failed: %v", err)
			respondJSONError(w, http.StatusUnauthorized, "Invalid or expired token")
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		ctx = context.WithValue(ctx, ContextKey("keycloak_claims"), claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole checks if the user has a specific role
func (k *KeycloakAuthMiddleware) RequireRole(role string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ContextKey("keycloak_claims")).(*auth.KeycloakClaims)
			if !ok {
				respondJSONError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			if !k.keycloak.HasRole(claims, role) {
				respondJSONError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole checks if the user has any of the specified roles
func (k *KeycloakAuthMiddleware) RequireAnyRole(roles ...string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ContextKey("keycloak_claims")).(*auth.KeycloakClaims)
			if !ok {
				respondJSONError(w, http.StatusUnauthorized, "Authentication required")
				return
			}

			hasRole := false
			for _, role := range roles {
				if k.keycloak.HasRole(claims, role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				respondJSONError(w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth validates the token if present, but doesn't require it
func (k *KeycloakAuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			next.ServeHTTP(w, r)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			next.ServeHTTP(w, r)
			return
		}

		claims, err := k.keycloak.ValidateToken(parts[1])
		if err != nil {
			// Invalid token, but we don't reject the request
			next.ServeHTTP(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		ctx = context.WithValue(ctx, ContextKey("keycloak_claims"), claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func respondJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write([]byte(`{"success":false,"error":"` + message + `"}`))
}
