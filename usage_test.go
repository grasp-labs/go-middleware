package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/grasp-labs/go-libs/aws/sqs"
	"github.com/grasp-labs/go-libs/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	productID = uuid.MustParse("40bb5b9b-0b3d-40f0-932f-2969200660d5")
)

func TestUsageConfig_toMiddleware(t *testing.T) {
	type fields struct {
		ProductID uuid.UUID
		MemoryMB  string
		sqsClient sqs.ClientSqs
	}
	tests := []struct {
		name       string
		fields     fields
		want       echo.MiddlewareFunc
		wantErrMsg string
		setup      func(f *fields)
		handler    func(c echo.Context) error
	}{
		{
			name: "ShouldErrorOnSQS",
			fields: fields{
				ProductID: productID,
				MemoryMB:  "1234",
			},
			setup: func(f *fields) {
				m := mocks.NewClientSqs(t)
				m.EXPECT().
					SendMsg(context.Background(), mock.Anything).
					Return(fmt.Errorf("foo sqs")).
					Once()
				f.sqsClient = m
			},
			wantErrMsg: "foo sqs",
			handler: func(c echo.Context) error {
				return nil
			},
		},
		{
			name: "ShouldErrorOnHandler",
			fields: fields{
				ProductID: productID,
				MemoryMB:  "1234",
			},
			setup: func(f *fields) {
				m := mocks.NewClientSqs(t)
				m.EXPECT().
					SendMsg(context.Background(), mock.Anything).
					Return(nil).
					Once()
				f.sqsClient = m
			},
			wantErrMsg: "foo handler",
			handler: func(c echo.Context) error {
				return fmt.Errorf("foo handler")
			},
		},
		{
			name: "ShouldLogUsage",
			fields: fields{
				ProductID: productID,
				MemoryMB:  "1234",
			},
			setup: func(f *fields) {
				m := mocks.NewClientSqs(t)
				m.EXPECT().
					SendMsg(context.Background(), mock.Anything).
					Return(nil).
					Once()
				f.sqsClient = m
			},
			handler: func(c echo.Context) error {
				return c.String(http.StatusOK, "Hello, World!")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(&tt.fields)

			c := &UsageConfig{
				ProductID: tt.fields.ProductID,
				MemoryMB:  tt.fields.MemoryMB,
			}

			h, err := c.toMiddleware(tt.fields.sqsClient)
			if err != nil {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			ctx := e.NewContext(req, rec)
			cc := &Context{
				Context:    ctx,
				TenantName: "foo_tenant",
				TenantID:   tenantID,
				RequestID:  requestID,
			}

			if err := h(tt.handler)(cc); err != nil {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
		})
	}
}
