package config

import (
	"gopkg.in/yaml.v3"
	"os"

	"github.com/pkg/errors"
)

type Http struct {
	//LogFile          string `yaml:"LogFile"`
	Port         int `yaml:"port"`
	WriteTimeout int `yaml:"write_timeout"`
	ReadTimeout  int `yaml:"read_timeout"`
}

type Mysql struct {
	Dsn         string `yaml:"dsn"`
	SlowLog     string `yaml:"slow_log"`
	SlowTime    int    `yaml:"slow_time"`
	MaxIdleConn int    `yaml:"max_idle_conn"`
	MaxOpenConn int    `yaml:"max_open_conn"`
}

type Postgres struct {
	//TODO
}

type Redis struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	Db       int    `yaml:"db"`
}

type Email struct {
	//TODO
}

type Config struct {
	Http     `yaml:"http"`
	Mysql    `yaml:"mysql"`
	Postgres `yaml:"postgres"`
	Redis    `yaml:"redis"`
	Email    `yaml:"email"`
}

const DefaultConfigFile = "config.yaml"

func Parse(file string) (*Config, error) {
	var cfg = &Config{}

	buf, err := os.ReadFile(file)
	if err != nil {
		return cfg, errors.Errorf("ReadInConfig: %v", err)
	}

	err = yaml.Unmarshal(buf, cfg)
	if err != nil {
		return cfg, errors.Errorf("ReadInConfig: %v", err)
	}

	if cfg.Port < 1024 {
		return cfg, errors.New("Please use a port between 1024~65535")
	}

	if cfg.Port > 65535 {
		return cfg, errors.New("The port number cannot exceed 65535")
	}

	return cfg, nil
}
