package server

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/gin-contrib/sessions"
	sessionsredis "github.com/gin-contrib/sessions/redis"
	"github.com/solunara/isb/src/model"
	"github.com/solunara/isb/src/model/xytmodel"
	"github.com/solunara/isb/src/repository"
	"github.com/solunara/isb/src/repository/cache"
	"github.com/solunara/isb/src/repository/dao"
	"github.com/solunara/isb/src/service"
	"github.com/solunara/isb/src/types/logger"
	"github.com/solunara/isb/src/web"
	"github.com/solunara/isb/src/web/middleware"
	"github.com/solunara/isb/src/web/xytweb"
	"go.uber.org/zap"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	gormLogger "gorm.io/gorm/logger"
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

func initLogger() logger.Logger {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}

func initDB(l logger.Logger) (*gorm.DB, error) {
	fmt.Println(viper.GetInt64("mysql.slow_time"))
	gorm_mysql, err := gorm.Open(
		mysql.Open(viper.GetString("mysql.dsn")),
		&gorm.Config{
			SkipDefaultTransaction:                   true,
			DisableForeignKeyConstraintWhenMigrating: true,
			NowFunc: func() time.Time {
				return time.Now().Local()
			},
			Logger: gormLogger.New(
				gormLoggerFunc(l.Debug),
				gormLogger.Config{
					// 慢查询阈值设置
					SlowThreshold:             time.Second,
					IgnoreRecordNotFoundError: false,
					ParameterizedQueries:      false,
					Colorful:                  false,
					LogLevel:                  gormLogger.Info,
				},
			),
		},
	)
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

func initMiddlewares(redisCmd redis.Cmdable, store sessionsredis.Store, l logger.Logger) []gin.HandlerFunc {
	bd := middleware.NewBuilder(func(ctx context.Context, al *middleware.AccessLog) {
		l.Debug("HTTP请求", logger.Field{Key: "al", Val: al})
	}).AllowReqBody(true).AllowRespBody()
	//viper.OnConfigChange(func(in fsnotify.Event) {
	//	ok := viper.GetBool("web.logreq")
	//	bd.AllowReqBody(ok)
	//})

	return []gin.HandlerFunc{
		corsHdl(),
		bd.Build(),
		sessions.Sessions("mysession", store),
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/users/signup").
			IgnorePaths("/users/login").
			Build(),
		//ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	return cors.New(cors.Config{
		//AllowCredentials: true,
		// 允许前端携带token字段
		AllowHeaders: []string{"Content-Type", "Authorization"},

		// 允许前端访问后端响应中带的头部
		ExposeHeaders: []string{"x-jwt-token"},

		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 24 * time.Hour,
	})
}

func initGinServer(mdls []gin.HandlerFunc) *gin.Engine {
	ginsrv := gin.Default()
	ginsrv.Use(mdls...)
	return ginsrv
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...any) {
	g(msg, logger.Field{Key: msg, Val: args})
}

func initRouters(ginEngine *gin.Engine, db *gorm.DB, cace redis.Cmdable) {
	// vbook-api
	userCache := cache.NewUserCache(cace)
	userDao := dao.NewUserDAO(db)
	userRepo := repository.NewUserRepository(userDao, userCache)
	userSrv := service.NewUserService(userRepo)
	userCtrl := web.NewUserHandler(userSrv)
	userCtrl.RegisterRoutes(ginEngine)

	// ms-api
	msGroup := ginEngine.Group("/ms")
	msUserDao := dao.NewMsUserDAO(db)
	msUserRepo := repository.NewMsUserRepository(msUserDao)
	msUserSrv := service.NewMsUserService(msUserRepo)
	msUserCtrl := web.NewMsUserHandler(msUserSrv)
	msUserCtrl.RegisterRoutes(msGroup)

	// xyt-api
	xytGroup := ginEngine.Group("/xyt")
	xytHospitalCtrl := xytweb.NewXytHospitalHandler(db)
	xytHospitalCtrl.RegisterRoutes(xytGroup)

	xytUserCtrl := xytweb.NewXytUserlHandler(cace, db)
	xytUserCtrl.RegisterRoutes(xytGroup)
}

func autoCreateTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.MsUser{},

		// xyt
		&xytmodel.Hospital{},
		&xytmodel.HospitalGrade{},
		&xytmodel.City{},
		&xytmodel.District{},
		&xytmodel.Department{},
		&xytmodel.Registration{},
		&xytmodel.Doctor{},
		&xytmodel.RegistrationType{},
		&xytmodel.Schedule{},
		&xytmodel.Patient{},

		&xytmodel.XytUser{},
	)
}
