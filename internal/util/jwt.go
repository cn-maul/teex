package util

import (
	"crypto/rand"
	"errors"
	"log"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTSecret is the signing key. It is read from the JWT_SECRET environment
// variable. If that variable is unset or empty, a random 32-byte key is
// generated and a warning is printed.
var JWTSecret []byte

func init() {
	if env := os.Getenv("JWT_SECRET"); env != "" {
		JWTSecret = []byte(env)
	} else {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			log.Fatalf("FATAL: failed to generate random JWT secret: %v", err)
		}
		JWTSecret = key
		log.Println("WARNING: JWT_SECRET environment variable is not set.")
		log.Println("  A random key has been generated for this session.")
		log.Println("  ALL users will be logged out when the server restarts.")
		log.Println("  Set JWT_SECRET in your environment to persist tokens across restarts.")
	}
}

type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT token
func GenerateToken(userID uint, username, role string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)), // 7 天过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(JWTSecret)
}

// ParseToken 解析 JWT token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return JWTSecret, nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, errors.New("invalid token")
}
