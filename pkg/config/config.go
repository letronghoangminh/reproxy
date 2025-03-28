package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Global GlobalConfig `mapstructure:"global"`
	Listeners []ListenerConfig `mapstructure:"listeners"`
}

type GlobalConfig struct {
	Port int `mapstructure:"port"`
	LogLevel string `mapstructure:"log_level"`
}

// TODO: validate host format
type ListenerConfig struct {
	Host []string `mapstructure:"host"`
	Handlers []HandlerConfig `mapstructure:"handlers"`
}

type HandlerConfig struct {
	Path string `mapstructure:"path"`
	StaticResponse StaticResponseConfig `mapstructure:"static_response"`
	StaticFiles StaticFilesConfig `mapstructure:"static_files"`
	ReverseProxy ReverseProxyConfig `mapstructure:"reverse_proxy"`
}

type StaticResponseConfig struct {
	StatusCode int `mapstructure:"status"`
	Body string `mapstructure:"body"`
}

type StaticFilesConfig struct {
	Root string `mapstructure:"root"`
	IndexFiles []string `mapstructure:"index_files"`
}

type ReverseProxyConfig struct {
	Upstreams []string `mapstructure:"upstreams"`
}

var (
	cfg *Config
)

func LoadConfig(path string) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetConfigFile(path)
	v.AutomaticEnv()
	v.AllowEmptyEnv(true)
	v.SetTypeByDefaultValue(true)

	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	for _, k := range v.AllKeys() {
		if value, ok := v.Get(k).(string); ok {
			v.Set(k, os.ExpandEnv(value))
		}
	}

	if err = v.Unmarshal(&cfg); err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
}

func GetConfig() *Config {
	if cfg == nil {
		panic("config is not loaded")
	}
	return cfg
}
