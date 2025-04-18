// Package config provides functionality to load and validate configuration files for the Reproxy application.
package config

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Global    GlobalConfig     `mapstructure:"global" validate:"required"`
	Listeners []ListenerConfig `mapstructure:"listeners" validate:"required,dive"`
}

type GlobalConfig struct {
	Port     int    `mapstructure:"port" validate:"required,gt=0,lt=65536"`
	LogLevel string `mapstructure:"log_level" validate:"required,oneof=debug info warn error fatal"`
}

type ListenerConfig struct {
	Host     []string        `mapstructure:"host" validate:"required,dive,hostname_port"`
	Handlers []HandlerConfig `mapstructure:"handlers" validate:"required,dive"`
}

type HandlerConfig struct {
	Matchers       MatchersConfig       `mapstructure:"matchers" validate:"omitempty"`
	StaticResponse StaticResponseConfig `mapstructure:"static_response"`
	StaticFiles    StaticFilesConfig    `mapstructure:"static_files"`
	ReverseProxy   ReverseProxyConfig   `mapstructure:"reverse_proxy"`
}

type MatchersConfig struct {
	Headers     map[string]string `mapstructure:"headers" validate:"omitempty,dive"`
	Query       map[string]string `mapstructure:"query" validate:"omitempty,dive"`
	Path        string            `mapstructure:"path" validate:"omitempty"`
	Method      []string          `mapstructure:"method" validate:"omitempty,dive,oneof=GET POST PUT DELETE PATCH OPTIONS HEAD *"`
	ClientCIDRs []string          `mapstructure:"client_cidrs" validate:"omitempty,dive,cidr"`
}

type StaticResponseConfig struct {
	StatusCode int    `mapstructure:"status" default:"200" validate:"omitempty,gte=100,lt=600"`
	Body       string `mapstructure:"body"`
}

type StaticFilesConfig struct {
	Root string `mapstructure:"root" validate:"omitempty,dir"`
}

type ReverseProxyConfig struct {
	Rewrite       string              `mapstructure:"rewrite" validate:"omitempty"`
	Upstreams     UpstreamConfig      `mapstructure:"upstreams" validate:"omitempty,required"`
	LoadBalancing LoadBalancingConfig `mapstructure:"load_balancing" validate:"omitempty"`
	AddHeaders    map[string]string   `mapstructure:"add_headers" validate:"omitempty,dive"`
	RemoveHeaders []string            `mapstructure:"remove_headers" validate:"omitempty,dive"`
}

type UpstreamConfig struct {
	Static  []string                `mapstructure:"static" validate:"omitempty"`
	Dynamic []DynamicUpstreamConfig `mapstructure:"dynamic" validate:"omitempty,dive"`
}

type DynamicUpstreamConfig struct {
	Type  string `mapstructure:"type" validate:"required,oneof=A AAAA CNAME"`
	Value string `mapstructure:"value" validate:"required"`
}

type LoadBalancingConfig struct {
	Strategy    string `mapstructure:"strategy" validate:"omitempty,oneof=round_robin random ip_hash uri_hash sticky"`
	Retries     int    `mapstructure:"retries" default:"3" validate:"omitempty,gte=0,lte=10"`
	TryInterval int    `mapstructure:"try_interval" default:"5" validate:"omitempty,gte=0,lte=60"`
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

	validate := validator.New(validator.WithRequiredStructEnabled())

	err = validate.Struct(cfg)
	if err != nil {
		var validateErrs validator.ValidationErrors
		if errors.As(err, &validateErrs) {
			fmt.Println("Configuration validation errors:")
			for _, e := range validateErrs {
				switch e.Tag() {
				case "required":
					fmt.Printf("  - %s is required but was not provided\n", e.Namespace())
				case "gt", "gte":
					fmt.Printf("  - %s must be greater than %s (got: %v)\n", e.Namespace(), e.Param(), e.Value())
				case "lt", "lte":
					fmt.Printf("  - %s must be less than %s (got: %v)\n", e.Namespace(), e.Param(), e.Value())
				case "oneof":
					fmt.Printf("  - %s must be one of [%s] (got: %v)\n", e.Namespace(), e.Param(), e.Value())
				case "url":
					fmt.Printf("  - %s must be a valid URL (got: %v)\n", e.Namespace(), e.Value())
				case "dir":
					fmt.Printf("  - %s must be a valid directory path (got: %v)\n", e.Namespace(), e.Value())
				case "hostname_port":
					fmt.Printf("  - %s must be a valid host:port combination (got: %v)\n", e.Namespace(), e.Value())
				default:
					fmt.Printf("  - %s failed validation: %s=%s (got: %v)\n", e.Namespace(), e.Tag(), e.Param(), e.Value())
				}
			}
		}
		log.Fatalf("config validation failed, see errors above")
	}
}

func GetConfig() *Config {
	if cfg == nil {
		panic("config is not loaded")
	}
	return cfg
}
