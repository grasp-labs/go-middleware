package main

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4/middleware"

	custommiddleware "github.com/grasp-labs/go-middleware"
)

func main() {
	// Create server
	e := echo.New()

	// Echo middlewares
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Create SSM client
	ssmClient, err := paramstore.NewClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Get param from ssm
	key, err := custommiddleware.GetJWTKey(context.Background(), ssmClient)
	if err != nil {
		log.Fatal(err)
	}

	// Use middlewares
	e.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(key),
		NewClaimsFunc: custommiddleware.NewClaimsFunction,
	}))
	e.Use(custommiddleware.NewCustomContextMiddleware)
	e.Use(custommiddleware.Dispatch(context.Background(), "audit-table"))
	e.Use(custommiddleware.UsageWithConfig(context.Background(), custommiddleware.UsageConfig{
		ProductID: uuid.New(),
		MemoryMB:  "1234",
	}))

	e.GET("/", func(c echo.Context) error {
		token, ok := c.Get("user").(*jwt.Token) // by default token is stored under `user` key
		if !ok {
			return errors.New("JWT token missing or invalid")
		}
		claims, ok := token.Claims.(*custommiddleware.JWTClaims) // by default claims is of type `jwt.MapClaims`
		if !ok {
			return errors.New("failed to cast claims as jwt.MapClaims")
		}
		return c.JSON(http.StatusOK, claims)
	})

	if err := e.Start(":8080"); err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
