package middleware

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/grasp-labs/go-libs/aws/sqs"
	"github.com/labstack/echo/v4"
)

// UsageConfig defines the config for UsageWithConfig middleware.
type UsageConfig struct {
	ProductID uuid.UUID
	MemoryMB  string
}

// UsageWithConfig returns a middleware for tracing service usage in api applications.
func UsageWithConfig(ctx context.Context, cfg UsageConfig, queueName ...string) echo.MiddlewareFunc {
	sqsQueueName := ""
	switch os.Getenv("BUILDING_MODE") {
	case "test":
		sqsQueueName = "daas-service-cost-handler-usage-queue-test"
	case "dev":
		sqsQueueName = "daas-service-cost-handler-usage-queue-dev"
	case "prod":
		sqsQueueName = "daas-service-cost-handler-usage-queue-prod"
	default:
		panic("unknown building mode!")
	}
	if len(queueName) != 0 {
		sqsQueueName = queueName[0]
	}
	client, err := sqs.NewClient(ctx, sqsQueueName)
	if err != nil {
		panic(err)
	}

	mw, err := cfg.toMiddleware(client)
	if err != nil {
		panic(err)
	}

	return mw
}

func (c *UsageConfig) toMiddleware(sqsClient sqs.ClientSqs) (echo.MiddlewareFunc, error) {
	if c.ProductID == uuid.Nil {
		return nil, fmt.Errorf("usage middleware - product id is nil")
	}
	if c.MemoryMB == "" {
		return nil, fmt.Errorf("usage middleware - memory MB is empty")
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {
			cc, ok := ctx.(*Context)
			if !ok {
				return fmt.Errorf("cannot cast context to custom context")
			}

			startTime := time.Now()
			handlerError := next(cc)
			processTime := time.Now()

			if cc.TenantID != uuid.Nil {
				queueInput := map[string]types.MessageAttributeValue{
					"product_id": {
						DataType:    aws.String("String"),
						StringValue: aws.String(c.ProductID.String()),
					},
					"tenant_id": {
						DataType:    aws.String("String"),
						StringValue: aws.String(cc.TenantID.String()),
					},
					"memory_mb": {
						DataType:    aws.String("String"),
						StringValue: aws.String(c.MemoryMB),
					},
					"start_timestamp": {
						DataType:    aws.String("String"),
						StringValue: aws.String(startTime.String()),
					},
					"end_timestamp": {
						DataType:    aws.String("String"),
						StringValue: aws.String(processTime.String()),
					},
					"workflow": {
						DataType:    aws.String("String"),
						StringValue: aws.String(cc.Request().URL.Path),
					},
				}
				if err := sqsClient.SendMsg(cc.Request().Context(), queueInput); err != nil {
					return err
				}
			}

			if handlerError != nil {
				return handlerError
			}
			return nil
		}
	}, nil
}
