package jwt

import (
	"errors"
	"fmt"
	"time"

	authpb "github.com/danielhoward314/packet-sentry/protogen/golang/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ClaimsType is an enum representing different claims to be used in JWT generation
type ClaimsType int

const (
	UnspecifiedClaim ClaimsType = iota
	AdminUISession
	APIAuthorization
)

// TokenType is an enum representing different token types
type TokenType int

const (
	UnspecifiedToken TokenType = iota
	Access
	Refresh
)

const (
	OrganizationIDKey            = "organization_id"
	AdministratorIDKey           = "administrator_id"
	AuthorizationRoleKey         = "authorization_role"
	TokenTypeKey                 = "token_type"
	ClaimsTypeKey                = "claims_type"
	issuerClaimValue             = "web-api.packet-sentry"
	webConsoleAudienceClaimValue = "web-console.packet-sentry"
	apiAudienceClaimValue        = "packet-sentry-api"
	TokenExpiredError            = "token expired"
	InvalidTokenError            = "invalid token"
)

const (
	adminUIAccessTokenExpiry           = 1 * time.Hour
	adminUIRefreshTokenExpiry          = 7 * 24 * time.Hour
	apiAuthorizationAccessTokenExpiry  = 15 * time.Minute
	apiAuthorizationRefreshTokenExpiry = 7 * 24 * time.Hour
)

type AdminUISessionClaims struct {
	OrganizationID    string `json:"organization_id"`
	AuthorizationRole string `json:"authorization_role"`
	jwt.RegisteredClaims
}

type APIAuthorizationClaims struct {
	OrganizationID    string `json:"organization_id"`
	AuthorizationRole string `json:"authorization_role"`
	jwt.RegisteredClaims
}

func GenerateJWT(secret string, tokenType TokenType, claimsType ClaimsType, claimsData map[string]interface{}) (string, error) {
	jwtID := uuid.NewString()
	// fields common across all claims types
	registeredClaims := jwt.RegisteredClaims{
		Issuer:   issuerClaimValue,
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ID:       jwtID,
	}
	switch claimsType {
	case AdminUISession:
		if tokenType == Access {
			registeredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(adminUIAccessTokenExpiry))
		} else {
			registeredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(adminUIRefreshTokenExpiry))
		}
		organizationID, ok := claimsData[OrganizationIDKey]
		if !ok {
			return "", errors.New("missing organization_id claims")
		}
		administratorID, ok := claimsData[AdministratorIDKey]
		if !ok {
			return "", errors.New("missing administrator_id claims")
		}
		authorizationRole, ok := claimsData[AuthorizationRoleKey]
		if !ok {
			return "", errors.New("missing authorization_role claims")
		}
		registeredClaims.Subject = administratorID.(string)
		registeredClaims.Audience = []string{webConsoleAudienceClaimValue}
		claims := &AdminUISessionClaims{
			OrganizationID:    organizationID.(string),
			AuthorizationRole: authorizationRole.(string),
			RegisteredClaims:  registeredClaims,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			return "", err
		}
		return tokenString, nil
	case APIAuthorization:
		if tokenType == Access {
			registeredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(apiAuthorizationAccessTokenExpiry))
		} else {
			registeredClaims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(apiAuthorizationRefreshTokenExpiry))
		}
		organizationID, ok := claimsData[OrganizationIDKey]
		if !ok {
			return "", errors.New("missing organization_id claims")
		}
		administratorID, ok := claimsData[AdministratorIDKey]
		if !ok {
			return "", errors.New("missing administrator_id claims")
		}
		authorizationRole, ok := claimsData[AuthorizationRoleKey]
		if !ok {
			return "", errors.New("missing authorization_role claims")
		}
		registeredClaims.Subject = administratorID.(string)
		registeredClaims.Audience = []string{apiAudienceClaimValue}
		claims := &APIAuthorizationClaims{
			OrganizationID:    organizationID.(string),
			AuthorizationRole: authorizationRole.(string),
			RegisteredClaims:  registeredClaims,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			return "", err
		}
		return tokenString, nil
	}
	return "", errors.New("unknown claims type")
}

func DecodeJWT(secret string, tokenString string, claimsType ClaimsType) error {
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	}

	switch claimsType {
	case AdminUISession:
		claims := &AdminUISessionClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
		if err != nil {
			return errors.New("failed to parse token")
		}
		if !token.Valid {
			return errors.New(InvalidTokenError)
		}
		if claims.ExpiresAt.Before(time.Now()) {
			return errors.New(TokenExpiredError)
		}
		return nil
	case APIAuthorization:
		claims := &APIAuthorizationClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
		if err != nil {
			return errors.New("failed to parse token")
		}
		if !token.Valid {
			return errors.New(InvalidTokenError)
		}
		if claims.ExpiresAt.Before(time.Now()) {
			return errors.New(TokenExpiredError)
		}
		return nil
	}
	return errors.New("unknown claims type")
}

func GetClaimsTypeFromProtoEnum(ct authpb.ClaimsType) (ClaimsType, error) {
	switch ct {
	case authpb.ClaimsType_ADMIN_UI_SESSION:
		return AdminUISession, nil
	case authpb.ClaimsType_API_AUTHORIZATION:
		return APIAuthorization, nil
	default:
		return AdminUISession, errors.New("invalid claims type")
	}
}
