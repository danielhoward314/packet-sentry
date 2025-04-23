package dao

import (
	psJWT "github.com/danielhoward314/packet-sentry/jwt"
)

type TokenData struct {
	OrganizationID    string           `json:"organization_id"`
	AdministratorID   string           `json:"administrator_id"`
	AuthorizationRole string           `json:"authorization_role"`
	TokenType         psJWT.TokenType  `json:"token_type"`
	ClaimsType        psJWT.ClaimsType `json:"claims_type"`
}

// TokenDatastore defines the interface for access token operations in a key-value datastore
type TokenDatastore interface {
	Create(tokenData *TokenData, tokenType psJWT.TokenType, claimsType psJWT.ClaimsType) (string, error)
	Read(token string) (*TokenData, error)
	Decode(tokenType psJWT.TokenType, tokenString string, claimsType psJWT.ClaimsType) error
	Delete(token string) error
}
