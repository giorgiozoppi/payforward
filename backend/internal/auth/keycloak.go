package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type KeycloakAuth struct {
	realm        string
	serverURL    string
	clientID     string
	clientSecret string
	publicKeys   map[string]*rsa.PublicKey
	mu           sync.RWMutex
}

type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type KeycloakClaims struct {
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
	PreferredUsername string `json:"preferred_username"`
	Name              string `json:"name"`
	GivenName         string `json:"given_name"`
	FamilyName        string `json:"family_name"`
	RealmAccess       struct {
		Roles []string `json:"roles"`
	} `json:"realm_access"`
	ResourceAccess map[string]struct {
		Roles []string `json:"roles"`
	} `json:"resource_access"`
	jwt.RegisteredClaims
}

func NewKeycloakAuth(serverURL, realm, clientID, clientSecret string) *KeycloakAuth {
	ka := &KeycloakAuth{
		realm:        realm,
		serverURL:    serverURL,
		clientID:     clientID,
		clientSecret: clientSecret,
		publicKeys:   make(map[string]*rsa.PublicKey),
	}

	// Load public keys on initialization
	go ka.refreshPublicKeys()

	// Refresh public keys periodically
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			ka.refreshPublicKeys()
		}
	}()

	return ka
}

func (ka *KeycloakAuth) refreshPublicKeys() error {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/certs", ka.serverURL, ka.realm)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks JWKSResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	ka.mu.Lock()
	defer ka.mu.Unlock()

	for _, key := range jwks.Keys {
		if key.Kty == "RSA" && key.Use == "sig" {
			pubKey, err := ka.parseRSAPublicKey(key)
			if err != nil {
				continue
			}
			ka.publicKeys[key.Kid] = pubKey
		}
	}

	return nil
}

func (ka *KeycloakAuth) parseRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, err
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

func (ka *KeycloakAuth) getPublicKey(kid string) (*rsa.PublicKey, error) {
	ka.mu.RLock()
	key, exists := ka.publicKeys[kid]
	ka.mu.RUnlock()

	if !exists {
		// Try to refresh keys
		if err := ka.refreshPublicKeys(); err != nil {
			return nil, fmt.Errorf("failed to refresh keys: %w", err)
		}

		ka.mu.RLock()
		key, exists = ka.publicKeys[kid]
		ka.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("key with kid %s not found", kid)
		}
	}

	return key, nil
}

func (ka *KeycloakAuth) ValidateToken(tokenString string) (*KeycloakClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &KeycloakClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key ID from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("kid not found in token header")
		}

		return ka.getPublicKey(kid)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*KeycloakClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims type")
	}

	// Validate issuer
	expectedIssuer := fmt.Sprintf("%s/realms/%s", ka.serverURL, ka.realm)
	if claims.Issuer != expectedIssuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", expectedIssuer, claims.Issuer)
	}

	// Validate audience
	if !ka.validateAudience(claims.Audience) {
		return nil, fmt.Errorf("invalid audience")
	}

	return claims, nil
}

func (ka *KeycloakAuth) validateAudience(audience jwt.ClaimStrings) bool {
	for _, aud := range audience {
		if aud == ka.clientID || aud == "account" {
			return true
		}
	}
	return false
}

func (ka *KeycloakAuth) ExtractBearerToken(r *http.Request) (string, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("authorization header not found")
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid authorization header format")
	}

	return parts[1], nil
}

func (ka *KeycloakAuth) GetUserInfo(ctx context.Context, accessToken string) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/userinfo", ka.serverURL, ka.realm)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info: status %d", resp.StatusCode)
	}

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return userInfo, nil
}

func (ka *KeycloakAuth) HasRole(claims *KeycloakClaims, role string) bool {
	// Check realm roles
	for _, r := range claims.RealmAccess.Roles {
		if r == role {
			return true
		}
	}

	// Check client roles
	if clientAccess, ok := claims.ResourceAccess[ka.clientID]; ok {
		for _, r := range clientAccess.Roles {
			if r == role {
				return true
			}
		}
	}

	return false
}
