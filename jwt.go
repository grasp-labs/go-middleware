package middleware

import (
	"context"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/grasp-labs/go-libs/aws/paramstore"
	"github.com/labstack/echo/v4"
)

type JWTClaims struct {
	jwt.RegisteredClaims
	Cls       string   `json:"cls"`
	Ver       string   `json:"ver"`
	Rol       []string `json:"rol"`
	Rsc       string   `json:"rsc"`
	TokenType string   `json:"token_type"`
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
