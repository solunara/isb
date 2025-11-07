package interceptor

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/solunara/isb/microgrpc"
	"google.golang.org/grpc"
)

var interceptor_server_first grpc.UnaryServerInterceptor = func(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {
	log.Println("这是第一个interceptor执行前")
	resp, err = handler(ctx, req)
	log.Println("这是第一个interceptor执行后")
	return
}

var interceptor_server_second grpc.UnaryServerInterceptor = func(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {
	log.Println("这是第二个interceptor执行前")
	resp, err = handler(ctx, req)
	log.Println("这是第二个interceptor执行后")
	return
}

func TestServer(t *testing.T) {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptor_server_first, interceptor_server_second),
	)
	// 优雅退出
	defer func() {
		server.GracefulStop()
	}()

	userServer := &microgrpc.Server{}
	microgrpc.RegisterUserServiceServer(server, userServer)
	l, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err)
	}
	err = server.Serve(l)
	t.Log(err)
}
