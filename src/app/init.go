package app

import (
	"context"
	"flag"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/solunara/isb/src/types/config"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"time"
)

var configFile *string

func init() {
	configFile = flag.String("c", "", "Path to the config file")
	flag.StringVar(configFile, "config", *configFile, "Path to the config file")
	flag.Parse()
}

func initConfig() error {
	if configFile != nil && *configFile != "" {
		if cfg, err := config.Parse(*configFile); err == nil {
			viper.Set("http.port", cfg.Http.Port)
			return nil
		}
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	return viper.ReadInConfig()
}

func initDB() (*gorm.DB, error) {
	gorm_mysql, err := gorm.Open(mysql.Open(viper.GetString("mysql.dsn")))
	if err != nil {
		return nil, err
	}
	sqlDb, err := gorm_mysql.DB()
	if err != nil {
		return nil, err
	}
	maxIdleConn := viper.GetInt("mysql.max_idle_conn")
	maxOpenConn := viper.GetInt("mysql.max_open_conn")
	if maxIdleConn > 0 {
		sqlDb.SetMaxIdleConns(maxIdleConn)
	}
	if maxOpenConn > 0 {
		sqlDb.SetMaxOpenConns(maxOpenConn)
	}
	return gorm_mysql, nil
}

func initRedis() (redis.Cmdable, error) {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})
	return redisCli, redisCli.Ping(context.Background()).Err()
}

func initGinServer() *gin.Engine {
	ginsrv := gin.Default()
	ginsrv.Use(
		cors.New(cors.Config{
			AllowAllOrigins: true,

			//AllowCredentials: true,
			// 允许前端携带token字段
			AllowHeaders: []string{"Content-Type", "Authorization"},

			// 允许前端访问后端响应中带的头部
			ExposeHeaders: []string{"x-jwt-token"},

			AllowOriginFunc: func(origin string) bool {
				return true
			},
			MaxAge: 24 * time.Hour,
		}),

		func(ctx *gin.Context) {
			println("add midddleware")
		},
	)
	return ginsrv
}
