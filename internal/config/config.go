package config

import (
	"balancer/pkg/logger"
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	Server      Port              `json:"server"`
	Backends    []BackendInfo     `json:"backends"`
	HealthCheck HealthCheckConfig `json:"healthCheck"`
	RateLim     RateLimiters      `json:"rateLimit"`
	Logger      *logger.Config    `json:"logger"`
}

type Port struct {
	Number int `json:"port"`
}

type BackendInfo struct {
	Url            string `json:"url"`
	MaxConnections int    `json:"maxConns"`
}

type HealthCheckConfig struct {
	Interval time.Duration `json:"interval"`
	Timeout  time.Duration `json:"timeout"`
	Endpoint string        `json:"endpoint"`
}

type RateLimiters struct {
	Cap  int           `json:"capacity"`
	Rate time.Duration `json:"rate"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config Config

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	if config.RateLim.Cap == 0 {
		config.RateLim.Cap = 100
	}
	if config.RateLim.Rate == 0 {
		config.RateLim.Rate = time.Minute
	}
	if config.HealthCheck.Interval == 0 {
		config.HealthCheck.Interval = 10 * time.Second
	}
	if config.HealthCheck.Timeout == 0 {
		config.HealthCheck.Timeout = 2 * time.Second
	}

	return &config, nil

}
