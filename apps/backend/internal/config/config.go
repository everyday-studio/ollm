package config

import (
	"fmt"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig `yaml:"app"`
	DB  DBConfig  `yaml:"db"`
}

type AppConfig struct {
	Env   string `yaml:"env"`
	Port  int    `yaml:"port"`
	Debug bool   `yaml:"debug"`
}

type DBConfig struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Port     int    `yaml:"port"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

func LoadConfig(env string) (*Config, error) {
	if err := godotenv.Load(".env"); err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}

	viper.SetConfigName(fmt.Sprintf("config.%s", env))
	viper.AddConfigPath("./config")
	viper.AddConfigPath("../../config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}
