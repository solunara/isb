package interceptor

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/solunara/isb/microgrpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var interceptor_client_first grpc.UnaryClientInterceptor = func(
	ctx context.Context,
	method string,
	req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	log.Println("这是第一个interceptor执行前")
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Println("这是第一个interceptor执行后")
	return err
}

var interceptor_client_second grpc.UnaryClientInterceptor = func(
	ctx context.Context,
	method string,
	req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {
	log.Println("这是第二个interceptor执行前")
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Println("这是第二个interceptor执行后")
	return err
}

func TestClient(t *testing.T) {
	cc, err := grpc.Dial(
		"localhost:8090",
		grpc.WithChainUnaryInterceptor(interceptor_client_first, interceptor_client_second),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := microgrpc.NewUserServiceClient(cc)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	resp, err := client.GetById(ctx, &microgrpc.GetByIdRequest{})
	assert.NoError(t, err)
	print(resp)
	t.Log(resp.User)
}
