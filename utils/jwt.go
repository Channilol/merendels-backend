package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Secret key per firmare i token (da spostare in env variable)
var jwtSecret = []byte("merendels-secret-2025")

// JWTClaims struttura del payload JWT
type JWTClaims struct {
	UserID int `json:"user_id"`
	Email string `json:"email"`
	RoleID *int `json:"role_id"`
	HierarchyLevel *int `json:"hierarchy_level"`
	jwt.RegisteredClaims
}

// GenerateToken crea un nuovo token JWT
func GenerateToken(userId int, email string, roleID *int, hierarchyLevel *int) (string, error) {
	// Claims con i dati utente + scadenza
	claims := JWTClaims{
		UserID: userId,
		Email: email,
		RoleID: roleID,
		HierarchyLevel: hierarchyLevel,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Issuer: "merendels-backend",
		},
	}

	// Creo il token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Firmo il token con la secret key
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken verifica e decodifica il token JWT
func ValidateToken(tokenString string) (*JWTClaims, error) {
	// parse + verifica token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (any, error) {
		// verifica che sia firmato HMAC
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return  nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Verifico che il token sia valido
	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	// estrae i claims
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ExtractTokenFromHeader estrae il token dall'header authorization
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header required")
	}

	// Check se il bearer é corretto
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:], nil
	}

	return "", errors.New("invalid authorization token format")
}