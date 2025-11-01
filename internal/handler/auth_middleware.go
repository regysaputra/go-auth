package handler

import (
	"auth/internal/usecase"
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Define a custom key type to avoid collisions in context
type contextKey string

const UserIDContextKey = contextKey("UserID")

// AuthMiddleware create a new Chi middleware for JWT authentication
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeError(w, http.StatusUnauthorized, ErrMissingAuthHeader.Error())
			return
		}

		// The header should be in the format "Bearer <token>"
		headerPorts := strings.Split(authHeader, " ")
		if len(headerPorts) != 2 || strings.ToLower(headerPorts[0]) != "bearer" {
			writeError(w, http.StatusUnauthorized, ErrMalformedAuthHeader.Error())
			return
		}

		tokenString := headerPorts[1]
		secretKey := []byte(os.Getenv("SECRET_KEY"))

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure the signing method is what we expect (HMAC)
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}

			return secretKey, nil
		})

		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, ErrInvalidToken.Error())
			return
		}

		// Extract the user ID from the token c laims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeError(w, http.StatusUnauthorized, ErrInvalidToken.Error())
			return
		}

		// The 'sub' in claim hold our user ID. JWT stores numbers as float64
		userIDFloat, ok := claims["sub"].(float64)
		if !ok {
			writeError(w, http.StatusUnauthorized, ErrInvalidToken.Error())
			return
		}

		// Convert float to integer
		userID := int64(userIDFloat)

		// Add user ID to request context
		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)

		// Call the next handler in the chain with the new context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserIDFromContext is a helper function to safely retrieve the user ID from the context
func GetUserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value(UserIDContextKey).(int64)
	if !ok {
		return 0, usecase.ErrInvalidCredentials
	}

	return userID, nil
}
