package middleware_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/grasp-labs/go-libs/aws/paramstore"
	"github.com/grasp-labs/go-middleware"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

func ExampleNewCustomContextMiddleware() {
	// Create server
	e := echo.New()

	// Echo middlewares
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	// Create SSM client
	ssmClient, err := paramstore.NewClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Get param from ssm
	key, err := middleware.GetJWTKey(context.Background(), ssmClient)
	if err != nil {
		log.Fatal(err)
	}

	// Use middlewares
	e.Use(echojwt.WithConfig(echojwt.Config{
		SigningKey:    []byte(key),
		NewClaimsFunc: middleware.NewClaimsFunction,
	}))
	e.Use(middleware.NewCustomContextMiddleware)

	e.GET("/", func(c echo.Context) error {
		token, ok := c.Get("user").(*jwt.Token) // by default token is stored under `user` key
		if !ok {
			return errors.New("JWT token missing or invalid")
		}
		claims, ok := token.Claims.(*middleware.JWTClaims) // by default claims is of type `jwt.MapClaims`
		if !ok {
			return errors.New("failed to cast claims as jwt.MapClaims")
		}
		return c.JSON(http.StatusOK, claims)
	})

	if err := e.Start(":8080"); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	// Output:
	// JWTClaims
}

func ExampleGetJWTKey() {
	// Create SSM client
	ssmClient, err := paramstore.NewClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Get param from ssm
	key, err := middleware.GetJWTKey(context.Background(), ssmClient)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(key)
	// Output:
	// 1234
}

func ExampleDispatch() {
	// Create server
	e := echo.New()

	// Echo middlewares
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	e.Use(middleware.Dispatch(context.Background(), "audit-dynamo-table"))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "hello world")
	})

	if err := e.Start(":8080"); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	// Output:
	// hello world
}

func ExampleUsageWithConfig() {
	// Create server
	e := echo.New()

	// Echo middlewares
	e.Use(echomiddleware.Logger())
	e.Use(echomiddleware.Recover())

	e.Use(middleware.UsageWithConfig(context.Background(), middleware.UsageConfig{
		ProductID: uuid.MustParse("Product UUID"),
		MemoryMB:  "1024",
	}))

	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "hello world")
	})

	if err := e.Start(":8080"); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	// Output:
	// hello world
}
