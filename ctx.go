package middleware

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Context struct {
	echo.Context
	Sub        string    `json:"sub"`
	Aud        []string  `json:"aud"`
	Rol        []string  `json:"rol"`
	Cls        string    `json:"cls"`
	Ver        string    `json:"ver"`
	TenantName string    `json:"tenant_name"`
	TenantID   uuid.UUID `json:"tenant_id"`
	RequestID  uuid.UUID `json:"request_id"`
}

func (c *Context) SetDataFromClaims(a JWTClaims) {
	c.Sub = a.Sub
	c.Aud = a.Aud
	c.Rol = a.Rol
	c.Cls = a.Cls
	c.Ver = a.Ver
	parts := strings.Split(a.Rsc, ":")
	if len(parts) == 2 {
		c.TenantID = uuid.MustParse(parts[0])
		c.TenantName = parts[1]
	}
}

func (c *Context) UserAndTenantIsPresent() bool {
	return c.TenantID != uuid.Nil && c.Sub != ""
}

// NewCustomContextMiddleware creates a middleware that enriches the echo context with custom data
// that can be found under context.Get("user").
// The middleware retrieves the JWT token from the context under the key "user" and validates it.
// It then extracts custom claims from the JWT token and sets them in the context.
// Additionally, it retrieves the Request-ID from the header and sets it in the context.
func NewCustomContextMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		token, ok := c.Get("user").(*jwt.Token) // by default token is stored under `user` key
		if !ok {
			return fmt.Errorf("JWT token missing or invalid")
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok {
			return fmt.Errorf("failed to cast claims as jwt.JWTClaims")
		}

		reqID := c.Response().Header().Get("X-Request-Id")
		if reqID == "" {
			return fmt.Errorf("failed to get Request-ID from header")
		}

		cc := &Context{
			Context:   c,
			RequestID: uuid.MustParse(reqID),
		}
		cc.SetDataFromClaims(*claims)

		return next(cc)
	}
}
