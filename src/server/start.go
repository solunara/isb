package server

import (
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"log"
)

func Start() {
	err := initConfig()
	if err != nil {
		log.Printf("[Err] init config: %v", err)
		log.Printf("[Warn] start by default...")
		StartDefault()
	}

	dbCli, err := initDB()
	if err != nil {
		log.Printf("[Err] init db client: %v", err)
		return
	}
	_ = dbCli

	redisCli, err := initRedis()
	if err != nil {
		log.Printf("[Err] init redis client: %v", err)
		return
	}
	_ = redisCli

	ginServer := initGinServer()

	ginServer.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "get request")
	})
	ginServer.POST("/", func(ctx *gin.Context) {
		ctx.String(200, "post request")
	})
	ginServer.PUT("/", func(ctx *gin.Context) {
		ctx.String(200, "put request")
	})
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
