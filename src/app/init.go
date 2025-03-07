package app

import (
	"context"
	"flag"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var configFile *string

func init() {
	configFile = flag.String("c", "", "Path to the config file")
	flag.StringVar(configFile, "config", *configFile, "Path to the config file")
	flag.Parse()
}

func initConfig() error {
	// first config is command parameter
	if configFile != nil && *configFile != "" {
		viper.AddConfigPath(*configFile)
		if err := viper.ReadInConfig(); err == nil {
			return nil
		}
	}

	// second config is local file
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err == nil {
		return nil
	}

	// third config is remote etcd
	err := viper.AddRemoteProvider("etcd3", "http://127.0.0.1:12379", "/isb")
	if err != nil {
		return err
	}
	viper.SetConfigType("yaml")
	return viper.ReadRemoteConfig()
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
			println("add middleware")
		},
	)
	return ginsrv
}
