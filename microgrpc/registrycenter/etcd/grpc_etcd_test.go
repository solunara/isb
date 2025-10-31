package etcd

import (
	"context"
	"net"
	"testing"
	"time"

	//"github.com/go-kit/kit/sd/etcdv3"

	"github.com/solunara/isb/microgrpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type EtcdTestSuit struct {
	suite.Suite
	client *etcdv3.Client
}

func (s *EtcdTestSuit) SetupSuite() {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(s.T(), err)
	s.client = client
}

func (s *EtcdTestSuit) TestClient() {
	bd, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)

	cc, err := grpc.NewClient(
		"etcd:///service/user",
		grpc.WithResolvers(bd),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(s.T(), err)

	client := microgrpc.NewUserServiceClient(cc)
	ctx, cancle := context.WithTimeout(context.Background(), time.Second)
	defer cancle()
	resp, err := client.GetById(ctx, &microgrpc.GetByIdRequest{
		Id: 123,
	})
	require.NoError(s.T(), err)

	s.T().Log(resp.User)
}
func (s *EtcdTestSuit) TestServer() {
	em, err := endpoints.NewManager(s.client, "servive/user")
	require.NoError(s.T(), err)
	ctx, cancle := context.WithTimeout(context.Background(), time.Second)
	defer cancle()
	addr := "127.0.0.1:8090"
	key := "service/user/" + addr

	l, err := net.Listen("tcp", ":8090")
	require.NoError(s.T(), err)

	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 租期，单位是秒
	var ttl int64 = 5
	leaseResp, err := s.client.Grant(ctx, ttl)
	require.NoError(s.T(), err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		// 续约
		ch, err1 := s.client.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(s.T(), err1)
		for kaResp := range ch {
			s.T().Log(kaResp.String())
		}
	}()

	go func() {
		// 模拟注册信息变动
		ticker := time.NewTicker(time.Second)
		for now := range ticker.C {
			ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
			err1 := em.Update(ctx1, []*endpoints.UpdateWithOpts{
				{
					Update: endpoints.Update{
						Op:  endpoints.Add,
						Key: key,
						Endpoint: endpoints.Endpoint{
							Addr:     addr,
							Metadata: now.String(),
						},
					},
					Opts: []etcdv3.OpOption{etcdv3.WithLease(leaseResp.ID)},
				},
				//{
				//	Update: endpoints.Update{
				//		Op:  endpoints.Delete,
				//		Key: key,
				//		Endpoint: endpoints.Endpoint{
				//			Addr:     addr,
				//			Metadata: now.String(),
				//		},
				//	},
				//},
			})
			// INSERT or update, save
			//err1 := em.AddEndpoint(ctx1, key, endpoints.Endpoint{
			//	Addr:     addr,
			//	Metadata: now.String(),
			//}, etcdv3.WithLease(leaseResp.ID))
			cancel1()
			if err1 != nil {
				s.T().Log(err1)
			}
		}
	}()

	server := grpc.NewServer()
	microgrpc.RegisterUserServiceServer(server, &microgrpc.Server{})
	server.Serve(l)
	kaCancel()
	err = em.DeleteEndpoint(ctx, key)
	if err != nil {
		s.T().Log(err)
	}
	server.GracefulStop()
	s.client.Close()
}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuit))
}
