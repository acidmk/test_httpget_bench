package config

import (
	"fmt"
	"github.com/spf13/viper"
	"sync"
)

type Config struct {
	Port            int  `mapstructure:"PORT"`
	Concurrency     int  `mapstructure:"CONCURRENCY"`
	RequestsPerHost int  `mapstructure:"REQUESTS_PER_HOST"`
	RequestTimeout  int  `mapstructure:"REQUEST_TIMEOUT"`
	UseHttpGet      bool `mapstructure:"USE_HTTP_GET"`
	ChunkSize       int  `mapstructure:"CHUNK_SIZE"`
}

var (
	getConfigOnce sync.Once
	mx            sync.RWMutex
	config        *Config
)

func GetConfig() *Config {
	getConfigOnce.Do(func() {
		mx.Lock()
		defer mx.Unlock()
		conf, err := loadConfig()
		config = conf

		if err != nil {
			panic(fmt.Sprintf("failed to load config: %s", err))
		}
	})

	mx.RLock()
	defer mx.RUnlock()
	return config
}

func loadConfig() (*Config, error) {
	var config *Config

	viper.SetDefault("Port", 8081)
	viper.SetDefault("Concurrency", 50000)
	viper.SetDefault("RequestsPerHost", 200)
	viper.SetDefault("RequestTimeout", 3)
	viper.SetDefault("UserHttpGet", true)
	viper.SetDefault("ChunkSize", 20)

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	if err := viper.ReadInConfig(); err != nil {
		return &Config{}, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return &Config{}, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}
