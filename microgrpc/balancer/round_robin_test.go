// grpc中使用轮询作为负载均衡算法
package balancer

import (
	"context"

	"time"

	"github.com/solunara/isb/microgrpc"
	"github.com/stretchr/testify/require"

	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *BalancerTestSuite) TestRRClient() {
	t := s.T()
	etcdResolver, err := resolver.NewBuilder(s.cli)
	require.NoError(s.T(), err)
	cc, err := grpc.NewClient("etcd:///service/user",
		grpc.WithResolvers(etcdResolver),
		grpc.WithDefaultServiceConfig(`
{
    "loadBalancingConfig": [
        {
            "round_robin": {}
        }
    ]
}
`),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := microgrpc.NewUserServiceClient(cc)
	for i := 0; i < 10; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		resp, err := client.GetById(ctx, &microgrpc.GetByIdRequest{Id: 123})
		cancel()
		require.NoError(t, err)
		t.Log(resp.User)
	}
}
