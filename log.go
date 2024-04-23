package middleware

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/grasp-labs/go-libs/aws/dynamodb"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type item struct {
	ID          uuid.UUID     `json:"id" dynamodbav:"id"`
	TenantID    uuid.UUID     `json:"tenant_id" dynamodbav:"tenant_id"`
	UserID      string        `json:"user_id" dynamodbav:"user_id"`
	URL         string        `json:"url" dynamodbav:"url"`
	Method      string        `json:"method" dynamodbav:"method"`
	ClientIP    string        `json:"client_ip" dynamodbav:"client_id"`
	StatusCode  int           `json:"status_code" dynamodbav:"status_code"`
	CreatedAt   time.Time     `json:"created_at" dynamodbav:"created_at"`
	ProcessTime time.Duration `json:"process_time" dynamodbav:"process_time"`
}

// Dispatch all requests made to api, internal and internet facing,
// a record will be created on which user, tenant and service, how the service was used,
// from which IP at what time.
func Dispatch(ctx context.Context, dynamodbTable string) echo.MiddlewareFunc {
	client, err := dynamodb.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	mw, err := toMiddleware(client, dynamodbTable)
	if err != nil {
		panic(err)
	}

	return mw
}

func toMiddleware(dynamo dynamodb.ClientDynamoDB, table string) (echo.MiddlewareFunc, error) {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc, ok := c.(*Context)
			if !ok {
				return fmt.Errorf("cannot parse context to custom context")
			}

			startTime := time.Now()
			handlerError := next(cc)

			processTime := time.Since(startTime)
			cc.Response().Header().Add("X-Process-Time", processTime.String())

			if ip := cc.Response().Header().Get("x-forwarded-for"); cc.UserAndTenantIsPresent() && ip != "" {
				inNetworks, err := checkIPInNetworks(ip)
				if err != nil {
					return err
				}

				if !inNetworks {
					auditItem := item{
						ID:          cc.RequestID,
						URL:         cc.Request().URL.String(),
						Method:      cc.Request().Method,
						ClientIP:    ip,
						StatusCode:  cc.Response().Status,
						TenantID:    cc.TenantID,
						UserID:      cc.UserID,
						CreatedAt:   startTime,
						ProcessTime: processTime,
					}

					if err := dynamo.PutItem(cc.Request().Context(), table, auditItem); err != nil {
						return err
					}
					log.Info(auditItem)
				}
			}
			if handlerError != nil {
				return handlerError
			}

			return nil
		}
	}, nil
}

func checkIPInNetworks(ip string) (bool, error) {
	networks := []string{
		"0.0.0.0/8",
		"10.0.0.0/8",
		"100.64.0.0/10",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"172.16.0.0/12",
		"172.0.0.0/8",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"192.88.99.0/24",
		"192.168.0.0/16",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"240.0.0.0/4",
		"255.255.255.255/32",
		"224.0.0.0/4",
	}

	ipAddr := net.ParseIP(ip)
	if ipAddr == nil {
		return false, nil
	}

	for _, network := range networks {
		_, ipNet, err := net.ParseCIDR(network)
		if err != nil {
			return false, err
		}

		if ipNet.Contains(ipAddr) {
			return true, nil
		}
	}

	return false, nil
}
