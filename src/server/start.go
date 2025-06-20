package server

import (
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	sessionsredis "github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func Start() {
	err := InitConfig(nil)
	if err != nil {
		log.Printf("[Err] init config: %v", err)
		log.Printf("[Warn] start by default...")
		StartDefault()
	}

	dbCli, err := InitDB(InitLogger())
	if err != nil {
		log.Printf("[Err] init db client: %v", err)
		return
	}

	err = autoCreateTable(dbCli)
	if err != nil {
		log.Printf("[Err] autoCreateTable: %v", err)
		return
	}

	redisCli, err := InitRedis()
	if err != nil {
		log.Printf("[Err] init redis client: %v", err)
		return
	}

	store, err := sessionsredis.NewStore(16,
		"tcp", viper.GetString("redis.addr"), "",
		[]byte("95osj3fUD7fo0mlYdDbncXz4VD2igvf0"), []byte("0Pf2r0wZBpXVXlQNdpwCXN4ncnlnZSc3"))

	if err != nil {
		log.Printf("[Err] sessionsredis.NewStore: %v", err)
		return
	}

	ginServer := InitGinServer(InitMiddlewares(redisCli, store, InitLogger()))

	InitRouters(ginServer, dbCli, redisCli)

	if err := ginServer.Run(fmt.Sprintf(":%v", viper.GetInt("http.port"))); err != nil {
		log.Fatal(err)
	}
}

func StartDefault() {
	server := gin.Default()
	server.Use(cors.Default())
	server.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "default response: get request")
	})
	server.POST("/", func(ctx *gin.Context) {
		ctx.String(200, "default response: post request")
	})
	server.PUT("/", func(ctx *gin.Context) {
		ctx.String(200, "default response: put request")
	})
	if err := server.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
