package middleware

import (
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

var (
	clientID   = uuid.MustParse("b225b665-bfb7-42e6-a794-3c12f166e146")
	userClaims = &JWTClaims{
		User: &User{
			GlobalAdmin:   true,
			CustomerAdmin: true,
			TenantName:    "Foo Tenant",
			UserID:        "foo@user.bar",
			TenantID:      tenantID,
		},
	}
	appClaims = &JWTClaims{
		App: &App{
			TenantName: "Foo Tenant",
			User:       "foo@user.bar",
			ClientID:   clientID,
			TenantID:   tenantID,
		},
	}
)

func TestNewCustomContextMiddleware(t *testing.T) {
	type args struct {
		assertion echo.HandlerFunc
	}
	tests := []struct {
		name       string
		args       args
		wantErrMsg string
		jwtToken   any
		requestID  string
	}{
		{
			name:       "ShouldFailOnParseToken",
			wantErrMsg: "JWT token missing or invalid",
			jwtToken:   "foo_bar",
		},
		{
			name: "ShouldFailOnParseClaims",
			jwtToken: &jwt.Token{
				Claims: jwt.MapClaims{},
			},
			args: args{
				assertion: func(c echo.Context) error {
					return nil
				},
			},
			wantErrMsg: "failed to cast claims as jwt.JWTClaims",
		},
		{
			name:      "ShouldFailOnGetRequestID",
			requestID: "",
			jwtToken: &jwt.Token{
				Claims: userClaims,
			},
			args: args{
				assertion: func(c echo.Context) error {
					return nil
				},
			},
			wantErrMsg: "failed to get Request-ID from header",
		},
		{
			name: "ShouldCreateCustomContextApp",
			jwtToken: &jwt.Token{
				Claims: appClaims,
			},
			requestID: requestID.String(),
			args: args{
				assertion: func(c echo.Context) error {
					cc, ok := c.(*Context)
					if !ok {
						log.Fatalln("cannot cast context to custom context")
					}

					cc.Context = nil // nil this. hard to mock, and it's not important
					want := &Context{
						IsGlobalAdminUser: false,
						IsCustomerAdmin:   false,
						UserID:            "foo@user.bar",
						TenantName:        "Foo Tenant",
						TenantID:          tenantID,
						RequestID:         requestID,
						AppID:             clientID,
					}

					assert.Equal(t, want, cc)
					return nil
				},
			},
		},
		{
			name:      "ShouldCreateCustomContextUser",
			requestID: requestID.String(),
			jwtToken: &jwt.Token{
				Claims: userClaims,
			},
			args: args{
				assertion: func(c echo.Context) error {
					cc, ok := c.(*Context)
					if !ok {
						log.Fatalln("cannot cast context to custom context")
					}

					cc.Context = nil // nil this. hard to mock, and its not important
					want := &Context{
						IsGlobalAdminUser: true,
						IsCustomerAdmin:   true,
						UserID:            "foo@user.bar",
						TenantName:        "Foo Tenant",
						TenantID:          tenantID,
						RequestID:         requestID,
					}

					assert.Equal(t, want, cc)
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			req.Header.Set("Request-ID", tt.requestID)
			ctx := e.NewContext(req, rec)

			ctx.Set("user", tt.jwtToken)

			h := NewCustomContextMiddleware(tt.args.assertion)
			err := h(ctx)
			if err != nil {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
		})
	}
}
