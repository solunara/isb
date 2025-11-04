package balancer

import (
	"context"
	"log"

	"github.com/solunara/isb/microgrpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type FailedServer struct {
	microgrpc.UnimplementedUserServiceServer
	Name string
}

func (s *FailedServer) GetByID(ctx context.Context, request *microgrpc.GetByIdRequest) (*microgrpc.GetByIdResponse, error) {
	log.Println("进来了 failover")
	return nil, status.Errorf(codes.Unavailable, "假装我被熔断了")
}
