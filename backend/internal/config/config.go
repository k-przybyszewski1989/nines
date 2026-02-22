package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
	Port   string
}

// Load reads configuration from a .env file (if present) and environment
// variables. Environment variables take precedence over the .env file.
func Load(envFile string) (*Config, error) {
	v := viper.New()

	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", "3306")
	v.SetDefault("DB_USER", "nines")
	v.SetDefault("DB_PASS", "nines")
	v.SetDefault("DB_NAME", "nines")
	v.SetDefault("PORT", "8080")

	if _, err := os.Stat(envFile); err == nil {
		v.SetConfigFile(envFile)
		v.SetConfigType("dotenv")
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("read config file: %w", err)
		}
	}

	v.AutomaticEnv()

	return &Config{
		DBHost: v.GetString("DB_HOST"),
		DBPort: v.GetString("DB_PORT"),
		DBUser: v.GetString("DB_USER"),
		DBPass: v.GetString("DB_PASS"),
		DBName: v.GetString("DB_NAME"),
		Port:   v.GetString("PORT"),
	}, nil
}

func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
}
