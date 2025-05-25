package websockets

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"

	psJWT "github.com/danielhoward314/packet-sentry/jwt"
)

const (
	logAttrValSvcName = "websocket"
)

type WebSocketHandler struct {
	accessTokenJWTSecret string
	logger               *slog.Logger
	Upgrader             *websocket.Upgrader
}

func NewWebSocketHandler(allowList []string, accessTokenJWTSecret string, baseLogger *slog.Logger) *WebSocketHandler {
	childLogger := baseLogger.With(slog.String("service", logAttrValSvcName))
	return &WebSocketHandler{
		accessTokenJWTSecret: accessTokenJWTSecret,
		logger:               childLogger,
		Upgrader:             newUpgrader(allowList),
	}
}

func newUpgrader(allowList []string) *websocket.Upgrader {
	return &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return false
			}
			for _, allowed := range allowList {
				if strings.TrimSpace(allowed) == origin {
					return true
				}
			}
			return false
		},
	}
}

func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}

	claims := &psJWT.APIAuthorizationClaims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(h.accessTokenJWTSecret), nil
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

	conn, err := h.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Info("upgrade failed", "error", err)
		return
	}
	defer conn.Close()

	h.logger.Info("WebSocket connection established")

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			h.logger.Info("read failed", "error", err)
			break
		}
		h.logger.Info("received", "message", message)
		conn.WriteMessage(mt, message)
	}
}
