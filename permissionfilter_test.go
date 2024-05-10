package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestPermissionFilterConfig_toMiddleware(t *testing.T) {
	type fields struct {
		Roles []string
	}
	tests := []struct {
		name            string
		fields          fields
		wantErr         string
		handler         func(c echo.Context) error
		testHttpHandler http.HandlerFunc
	}{
		{
			name: "ShouldErrorFromDecodeRole",
			fields: fields{
				Roles: defaultRoles,
			},
			handler: func(c echo.Context) error {
				return c.String(http.StatusOK, "hello world")
			},
			testHttpHandler: func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusOK)
				res.Write([]byte("foo_bar"))
			},
			wantErr: "code=500, message=invalid character 'o' in literal false (expecting 'a')",
		},
		{
			name: "ShouldErrorFromMissingRole",
			fields: fields{
				Roles: defaultRoles,
			},
			handler: func(c echo.Context) error {
				return c.String(http.StatusOK, "hello world")
			},
			testHttpHandler: func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusOK)
				res.Write([]byte("[{\"name\":\"service.workflow.user\"}]"))
			},
			wantErr: "code=403, message=user has not enough entitlements",
		},
		{
			name: "ShouldErrorFromHandler",
			fields: fields{
				Roles: defaultRoles,
			},
			handler: func(c echo.Context) error {
				return echo.NewHTTPError(http.StatusNotFound, "foo_error")
			},
			testHttpHandler: func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusOK)
				res.Write([]byte("[{\"name\":\"service.workflow.user\"}, {\"name\":\"service.workflow.admin\"}]"))
			},
			wantErr: "code=404, message=foo_error",
		},
		{
			name: "ShouldMatchRoles",
			fields: fields{
				Roles: defaultRoles,
			},
			handler: func(c echo.Context) error {
				return c.String(http.StatusOK, "Hello, World!")
			},
			testHttpHandler: func(res http.ResponseWriter, req *http.Request) {
				res.WriteHeader(http.StatusOK)
				res.Write([]byte("[{\"name\":\"service.workflow.user\"}, {\"name\":\"service.workflow.admin\"}]"))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := setupTestServer(tt.testHttpHandler)
			if !assert.NoError(t, err) {
				return
			}
			defer ts.Close()

			cfg := &PermissionFilterConfig{
				Roles: tt.fields.Roles,
				Url:   ts.URL,
			}
			h, err := cfg.toMiddleware()
			if err != nil {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer foo_token")

			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			cc := &Context{
				Context:    ctx,
				TenantName: "foo_tenant",
				TenantID:   tenantID,
			}

			if err := h(tt.handler)(cc); err != nil {
				assert.EqualError(t, err, tt.wantErr)
				return
			}
		})
	}
}

func setupTestServer(handler http.Handler) (*httptest.Server, error) {
	ts := httptest.NewServer(handler)

	return ts, nil
}
