package startup

import (
	"log"

	sessionsredis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/solunara/isb/src/server"
	"github.com/spf13/viper"
)

func InitServer(data []byte) (*gin.Engine, redis.Cmdable) {
	err := server.InitConfig(data)
	if err != nil {
		log.Fatalf("[Err] init config: %v", err)
	}

	dbCli, err := server.InitDB(server.InitLogger())
	if err != nil {
		log.Fatalf("[Err] init db client: %v", err)
	}

	redisCli, err := server.InitRedis()
	if err != nil {
		log.Fatalf("[Err] init redis client: %v", err)
	}

	store, err := sessionsredis.NewStore(16,
		"tcp", viper.GetString("redis.addr"), "",
		[]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"), []byte("0Pf2r0wZBpXVXlQNdpwCXN4ncnlnZSc3"))

	if err != nil {
		log.Fatalf("[Err] sessionsredis.NewStore: %v", err)
	}

	ginServer := server.InitGinServer(server.InitMiddlewares(redisCli, store, server.InitLogger()))

	server.InitRouters(ginServer, dbCli, redisCli)

	return ginServer, redisCli
}
