package middleware

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/grasp-labs/go-libs/aws/paramstore"
	"github.com/grasp-labs/go-libs/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	param = &ssm.GetParameterOutput{
		Parameter: &types.Parameter{Value: aws.String("1234")},
	}
)

func TestGetJWTKey(t *testing.T) {
	type args struct {
		c         context.Context
		ssmClient paramstore.SSMClient
	}
	tests := []struct {
		name       string
		args       args
		want       string
		wantErrMsg string
		setup      func(a *args)
	}{
		{
			name: "ShouldErrorOnWrongEnvVariable",
			args: args{
				c: context.Background(),
			},
			wantErrMsg: "unknown BUILDING_MODE env variable",
			setup: func(a *args) {
				if err := os.Setenv("BUILDING_MODE", "foo_var"); err != nil {
					log.Fatalln(err)
				}
			},
		},
		{
			name: "ShouldErrorOnGetJWTKey",
			args: args{
				c: context.Background(),
			},
			wantErrMsg: "foo error",
			setup: func(a *args) {
				if err := os.Setenv("BUILDING_MODE", "test"); err != nil {
					log.Fatalln(err)
				}
				ssmMock := mocks.NewSSMClient(t)

				ssmMock.
					EXPECT().
					GetParameter(a.c, JwtTestKey, true).
					Return(nil, fmt.Errorf("foo error")).
					Once()

				a.ssmClient = ssmMock
			},
		},
		{
			name: "ShouldGetJWTKeyOnDev",
			args: args{
				c: context.Background(),
			},
			want: "1234",
			setup: func(a *args) {
				if err := os.Setenv("BUILDING_MODE", "dev"); err != nil {
					log.Fatalln(err)
				}

				ssmMock := mocks.NewSSMClient(t)

				ssmMock.
					EXPECT().
					GetParameter(a.c, JwtDevKey, true).
					Return(param, nil).
					Once()

				a.ssmClient = ssmMock
			},
		},
		{
			name: "ShouldGetJWTKeyOnProd",
			args: args{
				c: context.Background(),
			},
			want: "1234",
			setup: func(a *args) {
				if err := os.Setenv("BUILDING_MODE", "prod"); err != nil {
					log.Fatalln(err)
				}

				ssmMock := mocks.NewSSMClient(t)

				ssmMock.
					EXPECT().
					GetParameter(a.c, JwtProdKey, true).
					Return(param, nil).
					Once()

				a.ssmClient = ssmMock
			},
		},
		{
			name: "ShouldGetJWTKeyOnTest",
			args: args{
				c: context.Background(),
			},
			want: "1234",
			setup: func(a *args) {
				if err := os.Setenv("BUILDING_MODE", "test"); err != nil {
					log.Fatalln(err)
				}

				ssmMock := mocks.NewSSMClient(t)

				ssmMock.
					EXPECT().
					GetParameter(a.c, JwtTestKey, true).
					Return(param, nil).
					Once()

				a.ssmClient = ssmMock
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(&tt.args)

			got, err := GetJWTKey(tt.args.c, tt.args.ssmClient)
			if err != nil {
				assert.EqualError(t, err, tt.wantErrMsg)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
