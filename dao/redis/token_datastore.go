package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/danielhoward314/packet-sentry/dao"
	psJWT "github.com/danielhoward314/packet-sentry/jwt"
)

const (
	// access token TTLs have a longer duration than the JWT expiration
	// to allow for detecting a valid, but expired access token
	adminUIAccessTokenTTL          = 3 * time.Hour
	apiAuthorizationAccessTokenTTL = 45 * time.Minute

	// refresh token TTLs match the JWT expiration
	adminUIRefreshTokenTTL          = 7 * 24 * time.Hour
	apiAuthorizationRefreshTokenTTL = 7 * 24 * time.Hour
)

type tokenDatastore struct {
	accessTokenJWTSecret  string
	refreshTokenJWTSecret string
	client                *redis.Client
}

// NewTokenDatastore returns a redis implementation for the accessToken key-value datastore
func NewTokenDatastore(client *redis.Client, accessTokenJWTSecret string, refreshTokenJWTSecret string) dao.TokenDatastore {
	return &tokenDatastore{
		client:                client,
		accessTokenJWTSecret:  accessTokenJWTSecret,
		refreshTokenJWTSecret: refreshTokenJWTSecret,
	}
}

// Create generates an access or refresh JWT with the provided claims type and a TTL for the claims type
func (tds *tokenDatastore) Create(td *dao.TokenData, tokenType psJWT.TokenType, claimsType psJWT.ClaimsType) (string, error) {
	var secret string
	var ttl time.Duration
	switch tokenType {
	case psJWT.Access:
		secret = tds.accessTokenJWTSecret
		if claimsType == psJWT.AdminUISession {
			ttl = adminUIAccessTokenTTL
		} else {
			ttl = apiAuthorizationAccessTokenTTL
		}
	case psJWT.Refresh:
		secret = tds.refreshTokenJWTSecret
		if claimsType == psJWT.AdminUISession {
			ttl = adminUIRefreshTokenTTL
		} else {
			ttl = apiAuthorizationRefreshTokenTTL
		}
	default:
		return "", errors.New("invalid token type when trying to create token")
	}
	if td.OrganizationID == "" {
		return "", errors.New("invalid organization id")
	}
	if td.AdministratorID == "" {
		return "", errors.New("invalid administrator id")
	}
	if td.AuthorizationRole == "" {
		return "", errors.New("invalid authorization role")
	}
	claimsData := make(map[string]interface{})
	claimsData[psJWT.OrganizationIDKey] = td.OrganizationID
	claimsData[psJWT.AdministratorIDKey] = td.AdministratorID
	claimsData[psJWT.AuthorizationRoleKey] = td.AuthorizationRole
	claimsData[psJWT.TokenTypeKey] = td.TokenType
	claimsData[psJWT.ClaimsTypeKey] = td.ClaimsType
	token, err := psJWT.GenerateJWT(secret, tokenType, claimsType, claimsData)
	if err != nil {
		return "", err
	}
	tokenJSON, err := json.Marshal(td)
	if err != nil {
		return "", err
	}
	atStatus := tds.client.Set(context.Background(), token, tokenJSON, ttl)
	if atStatus.Err() != nil {
		return "", err
	}
	return token, nil
}

func (tds *tokenDatastore) Read(tokenStr string) (*dao.TokenData, error) {
	tokenJSON, err := tds.client.Get(context.Background(), tokenStr).Result()
	if err != nil {
		return nil, err
	}
	var token dao.TokenData
	err = json.Unmarshal([]byte(tokenJSON), &token)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Decode parses a JWT using the secret for the given token type
func (tds *tokenDatastore) Decode(tokenType psJWT.TokenType, tokenString string, claimsType psJWT.ClaimsType) error {
	var secret string
	switch tokenType {
	case psJWT.Access:
		secret = tds.accessTokenJWTSecret
	case psJWT.Refresh:
		secret = tds.refreshTokenJWTSecret
	default:
		return errors.New("invalid token type when trying to decode")
	}
	return psJWT.DecodeJWT(secret, tokenString, claimsType)
}

// Delete deletes a token by key from the key-value store
func (tds *tokenDatastore) Delete(jwt string) error {
	cmdStatus := tds.client.Del(context.Background(), jwt)
	if cmdStatus.Err() != nil {
		return cmdStatus.Err()
	}
	return nil
}
