package middleware

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Context struct {
	echo.Context
	IsGlobalAdminUser bool      `json:"is_global_admin_user"`
	IsCustomerAdmin   bool      `json:"is_customer_admin"`
	UserID            string    `json:"user_id"`
	TenantName        string    `json:"tenant_name"`
	TenantID          uuid.UUID `json:"tenant_id"`
	AppID             uuid.UUID `json:"app_id"`
	RequestID         uuid.UUID `json:"request_id"`
}

func (c *Context) SetDataFromUser(u *User) {
	if u != nil {
		c.IsGlobalAdminUser = u.GlobalAdmin
		c.IsCustomerAdmin = u.CustomerAdmin
		c.UserID = u.UserID
		c.TenantID = u.TenantID
		c.TenantName = u.TenantName
	}
}

func (c *Context) SetDataFromApp(a *App) {
	if a != nil {
		c.IsGlobalAdminUser = false
		c.IsCustomerAdmin = false
		c.UserID = a.User
		c.TenantID = a.TenantID
		c.TenantName = a.TenantName
		c.AppID = a.ClientID
	}
}

func (c *Context) UserAndTenantIsPresent() bool {
	return c.TenantID != uuid.Nil && c.UserID != ""
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

		reqID := c.Request().Header.Get("Request-ID")
		if reqID == "" {
			return fmt.Errorf("failed to get Request-ID from header")
		}

		cc := &Context{
			Context:   c,
			RequestID: uuid.MustParse(reqID),
		}
		cc.SetDataFromUser(claims.User)
		cc.SetDataFromApp(claims.App)

		return next(cc)
	}
}
