package server

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	sessionsredis "github.com/gin-contrib/sessions/redis"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/solunara/isb/src/config"
	"github.com/solunara/isb/src/model"
	"github.com/solunara/isb/src/model/hllmodel"
	"github.com/solunara/isb/src/model/xytmodel"
	"github.com/solunara/isb/src/pkg/metric"
	"github.com/solunara/isb/src/pkg/ratelimit"
	"github.com/solunara/isb/src/repository"
	"github.com/solunara/isb/src/repository/cache"
	"github.com/solunara/isb/src/repository/dao"
	"github.com/solunara/isb/src/service"
	"github.com/solunara/isb/src/service/oauth2/wechat"
	"github.com/solunara/isb/src/service/sms/localsms"
	"github.com/solunara/isb/src/service/sms/ratelimitSms"
	"github.com/solunara/isb/src/types/logger"
	"github.com/solunara/isb/src/web"
	"github.com/solunara/isb/src/web/hllweb"
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

func InitConfig(data []byte) error {
	if data != nil {
		viper.SetConfigType("yaml")
		err := viper.ReadConfig(bytes.NewReader(data))
		if err == nil {
			return nil
		}
	}

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

func InitLogger() logger.Logger {
	l, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	return logger.NewZapLogger(l)
}

func InitDB(l logger.Logger) (*gorm.DB, error) {
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

func InitRedis() (redis.Cmdable, error) {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("redis.password"),
		DB:       viper.GetInt("redis.db"),
	})
	return redisCli, redisCli.Ping(context.Background()).Err()
}

func InitPrometheus() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", viper.GetInt("http.prometheus_port")), nil)
	}()
}

func InitMiddlewares(redisCmd redis.Cmdable, store sessionsredis.Store, l logger.Logger) []gin.HandlerFunc {
	bd := middleware.NewLogBuilder(func(ctx context.Context, al *middleware.AccessLog) {
		l.Debug("HTTP请求", logger.Field{Key: "al", Val: al})
	}).AllowReqBody(true).AllowRespBody()
	//viper.OnConfigChange(func(in fsnotify.Event) {
	//	ok := viper.GetBool("web.logreq")
	//	bd.AllowReqBody(ok)
	//})

	return []gin.HandlerFunc{
		corsHdl(),
		bd.Build(),
		(&metric.MiddlewareBuilder{
			Namespace: "solunara_isb",
			Subsystem: "vbook",
			Name:      "gin_http",
			Help:      "统计gin的http接口",
		}).Build(),
		sessions.Sessions("mysession", store),
		middleware.NewLoginJWTMiddlewareBuilder().
			IgnorePaths("/user/signup").
			IgnorePaths("/user/login/email").
			IgnorePaths("/user/oauth2/wechat/authurl").
			IgnorePaths("/user/oauth2/wechat/callback").
			IgnorePaths("/user/login/sms").
			IgnorePaths("/user/login/sms/send").
			IgnorePaths("/xyt/user/phone/code").
			IgnorePaths("/xyt/user/login/phone").
			IgnorePaths("/xyt/hos/list").
			IgnorePaths("/xyt/hos/grade").
			IgnorePaths("/xyt/hos/region").
			IgnorePaths("/xyt/hos/detail").
			IgnorePaths("/xyt/hos/department").
			IgnorePaths("/hll/user/login").
			Build(),
		//ratelimit.NewBuilder(redisClient, time.Second, 100).Build(),
	}
}

func corsHdl() gin.HandlerFunc {
	var cfg = cors.DefaultConfig()
	cfg.AddAllowHeaders(config.HTTTP_HEADER_AUTH)
	cfg.AllowOriginFunc = func(origin string) bool {
		return true
	}
	return cors.New(
		cfg,
	)

	// return cors.New(cors.Config{
	// 	//AllowCredentials: true,
	// 	// 允许前端携带token字段
	// 	AllowHeaders: []string{config.HTTTP_HEADER_AUTH},

	// 	AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "HEAD", "OPTIONS"},

	// 	// 允许前端访问后端响应中带的头部
	// 	ExposeHeaders: []string{config.HTTP_HEADER_TOKEN},

	// 	AllowOriginFunc: func(origin string) bool {
	// 		return true
	// 	},
	// 	MaxAge: 24 * time.Hour,
	// })
}

func InitGinServer(mdls []gin.HandlerFunc) *gin.Engine {
	ginsrv := gin.Default()
	ginsrv.Use(mdls...)
	return ginsrv
}

type gormLoggerFunc func(msg string, fields ...logger.Field)

func (g gormLoggerFunc) Printf(msg string, args ...any) {
	g(msg, logger.Field{Key: msg, Val: args})
}

func InitWechatService() wechat.Service {
	appId, ok := os.LookupEnv("WECHAT_APP_ID")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_ID ")
	}
	appKey, ok := os.LookupEnv("WECHAT_APP_SECRET")
	if !ok {
		panic("没有找到环境变量 WECHAT_APP_SECRET")
	}
	return wechat.NewOauth2WechatService(appId, appKey, http.DefaultClient)
}

func InitRouters(ginEngine *gin.Engine, db *gorm.DB, cace redis.Cmdable) {
	// vbook-api
	userCache := cache.NewUserCache(cace)
	codeCache := cache.NewCaptchaCache(cace)
	userDao := dao.NewUserDAO(db)
	userRepo := repository.NewUserRepository(userDao, userCache)
	codeRepo := repository.NewCaptchaRepository(codeCache)
	userSrv := service.NewUserService(userRepo)
	//smsSvc := localsms.NewService()
	ratelimitSmsSvc := ratelimitSms.NewRateLimitSMSService(localsms.NewService(), ratelimit.NewRedisSlideWindowLimit(cace, time.Second, 1000))
	codeSvc := service.NewCaptchaService(codeRepo, ratelimitSmsSvc, "000000")
	userCtrl := web.NewUserHandler(userSrv, codeSvc)
	userCtrl.RegisterRoutes(ginEngine)

	wechatSvc := InitWechatService()
	oauth2WechatCtrl := web.NewOAuth2WechatHandler(wechatSvc, userSrv)
	oauth2WechatCtrl.RegisterRoutes(ginEngine)

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

	xytCityCtrl := xytweb.NewXytCiteslHandler(db)
	xytCityCtrl.RegisterRoutes(xytGroup)

	// hll api
	hllGroup := ginEngine.Group("/hll")

	hllUserCtrl := hllweb.NewHllUserlHandler(cace, db)
	hllUserCtrl.RegisterRoutes(hllGroup)
}

func autoCreateTable(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.MsUser{},

		/* ---------------- xyt --------------- */
		// 医院信息表
		&xytmodel.Hospital{},
		&xytmodel.HospitalGrade{},
		&xytmodel.Department{},
		&xytmodel.Registration{},
		&xytmodel.Doctor{},
		&xytmodel.RegistrationType{},
		&xytmodel.Schedule{},
		&xytmodel.Patient{},
		&xytmodel.RegisterOrder{},

		// 城市表
		&xytmodel.Province{},
		&xytmodel.City{},
		&xytmodel.District{},

		// 用户表
		&xytmodel.XytUser{},

		/* ---------------- hllmgr --------------- */
		// 用户表
		&hllmodel.HllUser{},
	)
}

// 省 1级数据
type CitiesLevelOne struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// 省/市 2级数据
type CitiesLevelTwo struct {
	Code     string           `json:"code"`
	Name     string           `json:"name"`
	Children []CitiesLevelOne `json:"children"`
}

// 省/市/区县 3级数据
type CitiesLevelThree struct {
	Code     string           `json:"code"`
	Name     string           `json:"name"`
	Children []CitiesLevelTwo `json:"children"`
}

func ReadJSONInsertToDB(db *gorm.DB) {
	data, err := os.ReadFile("pca-code.json")
	if err != nil {
		panic(err)
	}

	var provinces []CitiesLevelThree
	if err := json.Unmarshal(data, &provinces); err != nil {
		panic(err)
	}

	var category uint8
	var abbr Abbreviation
	// now := time.Now().Unix()
	for _, p := range provinces {
		// 插入省
		category = categoryByProvinceName(p.Name)
		abbr = abbrByProvinceName(p.Name)
		err = db.Table(xytmodel.TableProvince).Create(&xytmodel.Province{
			Name:     p.Name,
			Abbr_zh:  abbr.Zh,
			Abbr_en:  abbr.En,
			Code:     p.Code,
			Category: category,
		}).Error
		if err != nil {
			panic(err)
		}
		for _, c := range p.Children {
			// 插入市
			// fmt.Printf("INSERT INTO cities (name, code, province_name, province_code, category, created_at, updated_at) VALUES ('%s', '%s', '%s', '%s', %d, %d, %d);\n",
			// 	c.Name, c.Code, p.Name, p.Code, category, now, now)
			err = db.Table(xytmodel.TableCity).Create(&xytmodel.City{
				Name:         c.Name,
				Code:         c.Code,
				ProvinceName: p.Name,
				ProvinceCode: p.Code,
				Category:     category,
			}).Error
			if err != nil {
				panic(err)
			}
			for _, d := range c.Children {
				// 插入区/县
				// fmt.Printf("INSERT INTO districts (name, code, city_name, city_code, created_at, updated_at) VALUES ('%s', '%s', '%s', '%s', %d, %d);\n",
				// 	d.Name, d.Code, c.Name, c.Code, now, now)
				err = db.Table(xytmodel.TableDistrict).Create(&xytmodel.District{
					Name:         d.Name,
					Code:         d.Code,
					CityName:     c.Name,
					CityCode:     c.Code,
					ProvinceName: p.Name,
					ProvinceCode: p.Code,
					Category:     category,
				}).Error
				if err != nil {
					panic(err)
				}
			}
		}
	}
}

// 行政区划分类
func categoryByProvinceName(name string) uint8 {
	switch name {
	case "北京市", "北京":
		return 3
	case "天津市", "天津":
		return 3
	case "上海市", "上海":
		return 3
	case "重庆市", "重庆":
		return 3
	}
	if strings.Contains(name, "自治区") {
		return 2
	}
	return 1
}

type Abbreviation struct {
	Zh string // 中文简称，例如 "冀"
	En string // 英文简称，例如 "HE"
}

func abbrByProvinceName(name string) Abbreviation {
	switch name {
	case "北京市":
		return Abbreviation{"京", "BJ"}
	case "天津市":
		return Abbreviation{"津", "TJ"}
	case "上海市":
		return Abbreviation{"沪", "SH"}
	case "重庆市":
		return Abbreviation{"渝", "CQ"}
	case "河北省":
		return Abbreviation{"冀", "HE"}
	case "山西省":
		return Abbreviation{"晋", "SX"}
	case "辽宁省":
		return Abbreviation{"辽", "LN"}
	case "吉林省":
		return Abbreviation{"吉", "JL"}
	case "黑龙江省":
		return Abbreviation{"黑", "HL"}
	case "江苏省":
		return Abbreviation{"苏", "JS"}
	case "浙江省":
		return Abbreviation{"浙", "ZJ"}
	case "安徽省":
		return Abbreviation{"皖", "AH"}
	case "福建省":
		return Abbreviation{"闽", "FJ"}
	case "江西省":
		return Abbreviation{"赣", "JX"}
	case "山东省":
		return Abbreviation{"鲁", "SD"}
	case "河南省":
		return Abbreviation{"豫", "HA"}
	case "湖北省":
		return Abbreviation{"鄂", "HB"}
	case "湖南省":
		return Abbreviation{"湘", "HN"}
	case "广东省":
		return Abbreviation{"粤", "GD"}
	case "海南省":
		return Abbreviation{"琼", "HI"}
	case "四川省":
		return Abbreviation{"川", "SC"}
	case "贵州省":
		return Abbreviation{"黔", "GZ"}
	case "云南省":
		return Abbreviation{"滇", "YN"}
	case "陕西省":
		return Abbreviation{"陕", "SN"}
	case "甘肃省":
		return Abbreviation{"甘", "GS"}
	case "青海省":
		return Abbreviation{"青", "QH"}
	case "台湾省":
		return Abbreviation{"台", "TW"}
	case "内蒙古自治区":
		return Abbreviation{"蒙", "NM"}
	case "广西壮族自治区":
		return Abbreviation{"桂", "GX"}
	case "西藏自治区":
		return Abbreviation{"藏", "XZ"}
	case "宁夏回族自治区":
		return Abbreviation{"宁", "NX"}
	case "新疆维吾尔自治区":
		return Abbreviation{"新", "XJ"}
	case "香港特别行政区":
		return Abbreviation{"港", "HK"}
	case "澳门特别行政区":
		return Abbreviation{"澳", "MO"}
	default:
		return Abbreviation{"", ""}
	}
}
