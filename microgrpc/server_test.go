package microgrpc

import (
	"net"
	"testing"

	"google.golang.org/grpc"
)

func TestServer(t *testing.T) {
	server := grpc.NewServer()
	userServer := &Server{}
	RegisterUserServiceServer(server, userServer)

	l, err := net.Listen("tcp", ":8090")
	if err != nil {
		panic(err)
	}
	err = server.Serve(l)
	t.Log(err)
}
