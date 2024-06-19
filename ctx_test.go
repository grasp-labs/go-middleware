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
	clientID = uuid.MustParse("b225b665-bfb7-42e6-a794-3c12f166e146")
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
				Claims: &JWTClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "mock-sub",
						Issuer:  "mock-iss",
					},
				},
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
				Claims: &JWTClaims{
					RegisteredClaims: jwt.RegisteredClaims{
						Subject: "mock-sub",
					},
					Rsc: "b9db1d4a-4364-4452-a2df-fcd44f38a63b:mock-tenant-name",
				},
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
						TenantID:   uuid.MustParse("b9db1d4a-4364-4452-a2df-fcd44f38a63b"),
						TenantName: "mock-tenant-name",
						Sub:        "mock-sub",
						RequestID:  requestID,
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

			rec.Header().Set("X-Request-Id", tt.requestID)
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
