package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

var defaultRoles = []string{"service.workflow.user", "service.workflow.admin"}

type PermissionFilterConfig struct {
	Roles []string // Optional
	Url   string   // Optional
}

func PermissionFilterWithConfig(cfg PermissionFilterConfig) echo.MiddlewareFunc {
	if len(cfg.Roles) == 0 {
		cfg.Roles = defaultRoles
	}

	if cfg.Url == "" {
		switch os.Getenv("BUILDING_MODE") {
		case "dev":
			cfg.Url = "https://grasp-daas.com/api/entitlements-dev/v1/groups/"
		case "prod":
			cfg.Url = "https://grasp-daas.com/api/entitlements/v1/groups/"
		}
	}

	mw, err := cfg.toMiddleware()
	if err != nil {
		panic(err)
	}

	return mw
}

func (p *PermissionFilterConfig) toMiddleware() (echo.MiddlewareFunc, error) {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc, ok := c.(*Context)
			if !ok {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("cannot cast context to custom context"))
			}

			req, err := http.NewRequest(http.MethodGet, p.Url, nil)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}

			authToken, tenantID := cc.Request().Header.Get("Authorization"), cc.TenantID.String()
			if authToken == "" || tenantID == "" {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid Authorization header"))
			}

			req.Header.Set("Authorization", authToken)
			req.Header.Set("tenant-id", tenantID)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}

			defer resp.Body.Close()

			var result []struct {
				Name string `json:"name"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, err)
			}

			roles := make(map[string]struct{})
			for _, r := range result {
				roles[r.Name] = struct{}{}
			}

			rolesMatched := 0
			for _, r := range p.Roles {
				if _, ok := roles[r]; ok {
					rolesMatched++
				}
			}

			if rolesMatched != len(p.Roles) {
				return echo.NewHTTPError(http.StatusForbidden, fmt.Errorf("user has not enough entitlements"))
			}

			return next(cc)
		}
	}, nil
}
