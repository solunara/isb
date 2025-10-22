package grpc

import (
	"context"

	intrv1 "github.com/solunara/isb/api/proto/gen/intr/v1"
	"github.com/solunara/isb/interactive/service"
)

type InteractiveServiceServer struct {
	intrv1.UnimplementedInteractiveServiceServer
	// 核心业务逻辑一定是在 service 里面的
	svc service.InteractiveService
}

// CancelLike implements intrv1.InteractiveServiceServer.
func (i *InteractiveServiceServer) CancelLike(context.Context, *intrv1.CancelLikeRequest) (*intrv1.CancelLikeResponse, error) {
	panic("unimplemented")
}

// Collect implements intrv1.InteractiveServiceServer.
func (i *InteractiveServiceServer) Collect(context.Context, *intrv1.CollectRequest) (*intrv1.CollectResponse, error) {
	panic("unimplemented")
}

// Get implements intrv1.InteractiveServiceServer.
func (i *InteractiveServiceServer) Get(context.Context, *intrv1.GetRequest) (*intrv1.GetResponse, error) {
	panic("unimplemented")
}

// GetByIds implements intrv1.InteractiveServiceServer.
func (i *InteractiveServiceServer) GetByIds(context.Context, *intrv1.GetByIdsRequest) (*intrv1.GetByIdsResponse, error) {
	panic("unimplemented")
}

// IncrReadCnt implements intrv1.InteractiveServiceServer.
func (i *InteractiveServiceServer) IncrReadCnt(context.Context, *intrv1.IncrReadCntRequest) (*intrv1.IncrReadCntResponse, error) {
	panic("unimplemented")
}

// Like implements intrv1.InteractiveServiceServer.
func (i *InteractiveServiceServer) Like(context.Context, *intrv1.LikeRequest) (*intrv1.LikeResponse, error) {
	panic("unimplemented")
}

// mustEmbedUnimplementedInteractiveServiceServer implements intrv1.InteractiveServiceServer.
func (i *InteractiveServiceServer) mustEmbedUnimplementedInteractiveServiceServer() {
	panic("unimplemented")
}

var _ intrv1.InteractiveServiceServer = (*InteractiveServiceServer)(nil)
