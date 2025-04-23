package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt/v5"

	psPostgres "github.com/danielhoward314/packet-sentry/dao/postgres"
	psJWT "github.com/danielhoward314/packet-sentry/jwt"
)

type authMiddleware struct {
	redisClient               *redis.Client
	accessTokenJWTSecret      string
	pathsWithoutAuthorization []string
	primaryAdminPaths         []string
}

type AuthMiddleWare interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler)
}

func NewAuthMiddleware(redisClient *redis.Client, accessTokenJWTSecret string, pathsWithoutAuthorization []string, primaryAdminPaths []string) AuthMiddleWare {
	return &authMiddleware{
		redisClient:               redisClient,
		accessTokenJWTSecret:      accessTokenJWTSecret,
		pathsWithoutAuthorization: pathsWithoutAuthorization,
		primaryAdminPaths:         primaryAdminPaths,
	}
}

// ServeHTTP implements the API authorization middleware
// returning 401 status for any error to trigger the admin UI's axios interceptor
// to request a new access token
func (am *authMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
	if am.isUnprotectedPath(r.URL.Path) {
		next.ServeHTTP(w, r)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "authorization header required", http.StatusUnauthorized)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenString == "" {
		http.Error(w, "bearer token value required", http.StatusUnauthorized)
		return
	}

	ctx := context.Background()
	_, err := am.redisClient.Get(ctx, tokenString).Result()
	if err != nil {
		http.Error(w, "API access token not found", http.StatusUnauthorized)
		return
	}
	claims := &psJWT.APIAuthorizationClaims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(am.accessTokenJWTSecret), nil
	})
	if err != nil {
		http.Error(w, "failed to parse API access token", http.StatusUnauthorized)
		return
	}
	if !parsedToken.Valid {
		http.Error(w, "invalid API access token", http.StatusUnauthorized)
		return
	}
	expirationTime, err := claims.GetExpirationTime()
	if err != nil {
		http.Error(w, "failed to get expiration time from token", http.StatusUnauthorized)
		return
	}
	if expirationTime.Before(time.Now()) {
		http.Error(w, "expired API access token", http.StatusUnauthorized)
		return
	}
	isAuthorizedForResource := am.isAuthorizedForResource(r.URL.Path, claims.AuthorizationRole)
	if !isAuthorizedForResource {
		msg := fmt.Sprintf("invalid authorization role %s for path %s", claims.AuthorizationRole, r.URL.Path)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}
	next.ServeHTTP(w, r)
}

func (am *authMiddleware) isUnprotectedPath(path string) bool {
	for _, p := range am.pathsWithoutAuthorization {
		if p == path {
			return true
		}
	}
	return false
}

func (am *authMiddleware) isAuthorizedForResource(path string, authorizationRole string) bool {
	for _, primaryAdminPath := range am.primaryAdminPaths {
		isPrimaryAdminPath := strings.Contains(path, primaryAdminPath)
		if isPrimaryAdminPath && authorizationRole != psPostgres.PrimaryAdmin {
			return false
		}
	}
	return true
}
