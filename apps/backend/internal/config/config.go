package config

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App    AppConfig    `mapstructure:"app"`
	DB     DBConfig     `mapstructure:"db"`
	Secure SecureConfig `mapstructure:"secure"`
}

type AppConfig struct {
	Env      string `mapstructure:"env"`
	Port     int    `mapstructure:"port"`
	Debug    bool   `mapstructure:"debug"`
	LogLevel string `mapstructure:"log_level"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type SecureConfig struct {
	CORSAllowOrigins []string  `mapstructure:"cors_allow_origins"`
	JWT              JWTConfig `mapstructure:"jwt"`
}

type JWTConfig struct {
	PrivateKey           string       `mapstructure:"private_key_base64"`
	PublicKey            string       `mapstructure:"public_key_base64"`
	AccessExpirationMin  int          `mapstructure:"access_expiration_min"`
	RefreshExpirationDay int          `mapstructure:"refresh_expiration_day"`
	Cookie               CookieConfig `mapstructure:"cookie"`
}

type CookieConfig struct {
	Secure   bool   `mapstructure:"secure"`
	HTTPOnly bool   `mapstructure:"http_only"`
	SameSite string `mapstructure:"same_site"`
	Domain   string `mapstructure:"domain"`
}

func LoadConfig(env string) (*Config, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}

	currentDir := filepath.Dir(filename)
	projectRoot := filepath.Join(currentDir, "../..")
	configPath := filepath.Join(projectRoot, "config")
	envPath := filepath.Join(projectRoot, ".env")

	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}

	viper.SetConfigName(fmt.Sprintf("config.%s", env))
	viper.AddConfigPath(configPath)
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
