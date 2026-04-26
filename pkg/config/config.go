package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Port    string        `yaml:"port"`
	MongoDB MongoDBConfig `yaml:"mongodb"`
	Kafka   KafkaConfig   `yaml:"kafka"`
	Auth    AuthConfig    `yaml:"auth"`
	GRPC    GRPCConfig    `yaml:"grpc"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

type MongoDBConfig struct {
	URI                string `yaml:"uri"`
	Database           string `yaml:"database"`
	Collection         string `yaml:"collection"`
	MaxPoolSize        uint64 `yaml:"max_pool_size"`
	MinPoolSize        uint64 `yaml:"min_pool_size"`
	MaxConnIdleTimeSec int    `yaml:"max_conn_idle_time_second"`
}

type AuthConfig struct {
	JWTSecret        string `yaml:"jwt_secret"`
	JWTExpirationMin int    `yaml:"jwt_expiration_minutes"`
	UserCollection   string `yaml:"user_collection"`
}

type GRPCConfig struct {
	Port string `yaml:"port"`
}

func Read() *AppConfig {
	wd, er := os.Getwd()
	if er != nil {
		panic(fmt.Errorf("unable to get working directory, %v", er))
	}
	viper.SetConfigName("config")       // name of config file (without extension)
	viper.SetConfigType("yaml")         // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(wd + "/config") // call multiple times to add many search paths
	viper.AddConfigPath(".")            // optionally look for config in the working directory
	err := viper.ReadInConfig()         // Find and read the config file
	if err != nil {                     // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	var config AppConfig
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("unable to decode into struct, %v", err))
	}

	return &config
}
