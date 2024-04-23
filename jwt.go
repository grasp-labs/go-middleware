package middleware

import (
	"context"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/grasp-labs/go-libs/aws/paramstore"
	"github.com/labstack/echo/v4"
)

type User struct {
	GlobalAdmin   bool      `json:"global_admin"`
	CustomerAdmin bool      `json:"customer_admin"`
	TenantName    string    `json:"tenant_name"`
	UserID        string    `json:"user_id"`
	TenantID      uuid.UUID `json:"tenant_id"`
}

type App struct {
	TenantName string    `json:"tenant_name"`
	User       string    `json:"user"`
	ClientID   uuid.UUID `json:"client_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
}

type JWTClaims struct {
	jwt.RegisteredClaims
	User      *User  `json:"user"`
	App       *App   `json:"app"`
	Exp       int    `json:"exp"`
	Iat       int    `json:"iat"`
	UserID    int    `json:"user_id"`
	TokenType string `json:"token_type"`
	Jti       string `json:"jti"`
	Iss       string `json:"iss"`
	ExpiresIn string `json:"expires_in"`
}

const (
	JwtProdKey = "AUTH_JWT_PUBLIC_KEY_PROD"
	JwtDevKey  = "AUTH_JWT_PUBLIC_KEY_DEV"
	JwtTestKey = "AUTH_JWT_PUBLIC_KEY_TEST"
)

func NewClaimsFunction(_ echo.Context) jwt.Claims {
	claims := &JWTClaims{}
	return claims
}

// GetJWTKey Get JWT secret from AWS parameter store.
func GetJWTKey(c context.Context, ssmClient paramstore.SSMClient) (string, error) {
	jwtKey := ""
	switch os.Getenv("BUILDING_MODE") {
	case "test":
		jwtKey = JwtTestKey
	case "dev":
		jwtKey = JwtDevKey
	case "prod":
		jwtKey = JwtProdKey
	default:
		return "", fmt.Errorf("unknown BUILDING_MODE env variable")
	}

	param, err := ssmClient.GetParameter(c, jwtKey, true)
	if err != nil {
		return "", err
	}

	return *param.Parameter.Value, nil
}
