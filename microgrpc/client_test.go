package microgrpc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestClient(t *testing.T) {
	cc, err := grpc.Dial(":8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := NewUserServiceClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetById(ctx, &GetByIdRequest{})
	assert.NoError(t, err)
	print(resp)
	t.Log(resp.User)
}
