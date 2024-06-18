package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/grasp-labs/go-libs/aws/dynamodb"
	"github.com/grasp-labs/go-libs/mocks"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	tenantID  = uuid.MustParse("dd49bb44-ac56-4e70-8697-89603f4125f2")
	requestID = uuid.MustParse("03a3e1d6-8bf8-42ce-8e65-a83b2ba68c3a")
	userID    = "foo@bar.com"
)

func TestDispatch(t *testing.T) {
	type args struct {
		next echo.HandlerFunc
	}
	type fields struct {
		dynamoClient dynamodb.ClientDynamoDB
	}
	tests := []struct {
		name       string
		args       args
		fields     fields
		wantErrMsg string
		setup      func(f *fields)
	}{
		{
			name: "ShouldErrorFromHandler",
			args: args{
				next: func(c echo.Context) error {
					return fmt.Errorf("foo error")
				},
			},
			wantErrMsg: "foo error",
			setup: func(f *fields) {
				dbMock := mocks.NewClientDynamoDB(t)
				dbMock.EXPECT().
					PutItem(context.Background(), "dynamo-table", mock.Anything).
					Return(nil).
					Once()

				f.dynamoClient = dbMock
			},
		},
		{
			name: "ShouldErrorOnSaveDynamo",
			args: args{
				next: func(c echo.Context) error {
					return nil
				},
			},
			wantErrMsg: "foo dynamo",
			setup: func(f *fields) {
				dbMock := mocks.NewClientDynamoDB(t)
				dbMock.EXPECT().
					PutItem(context.Background(), "dynamo-table", mock.Anything).
					Return(fmt.Errorf("foo dynamo")).
					Once()

				f.dynamoClient = dbMock
			},
		},
		{
			name: "ShouldLogAudit",
			args: args{
				next: func(c echo.Context) error {
					return nil
				},
			},
			setup: func(f *fields) {
				dbMock := mocks.NewClientDynamoDB(t)
				dbMock.EXPECT().
					PutItem(context.Background(), "dynamo-table", mock.Anything).
					Return(nil).
					Once()

				f.dynamoClient = dbMock
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(&tt.fields)

			h, err := toMiddleware(tt.fields.dynamoClient, "dynamo-table")
			if err != nil {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}

			e := echo.New()

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()

			rec.Header().Set("x-forwarded-for", "1.1.1.1")
			ctx := e.NewContext(req, rec)

			cc := &Context{
				Context:   ctx,
				RequestID: requestID,
				TenantID:  tenantID,
				Sub:       userID,
			}
			if err := h(tt.args.next)(cc); err != nil {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.NotNil(t, rec.Header().Get("X-Process-Time"))
		})
	}
}

func TestCheckIPInNetworks(t *testing.T) {
	type args struct {
		ip string
	}
	tests := []struct {
		name       string
		args       args
		want       bool
		wantErrMsg string
	}{
		{
			name: "ShouldNotParse",
			args: args{
				"foo_bar",
			},
			want: false,
		},
		{
			name: "ShouldNotFindInNetwork",
			args: args{
				"192.50.2.1",
			},
			want: false,
		},
		{
			name: "ShouldFindInNetwork",
			args: args{
				"192.0.2.1",
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkIPInNetworks(tt.args.ip)
			if err != nil {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
