package postgres

import (
	"fmt"
	"time"
)

// Config представляет конфигурацию подключения к PostgreSQL
type Config struct {
	Host            string        `yaml:"host" env:"DB_HOST" default:"localhost"`
	Port            int           `yaml:"port" env:"DB_PORT" default:"5432"`
	User            string        `yaml:"user" env:"DB_USER" default:"postgres"`
	Password        string        `yaml:"password" env:"DB_PASSWORD" default:""`
	Database        string        `yaml:"database" env:"DB_NAME" default:"gophkeeper"`
	SSLMode         string        `yaml:"ssl_mode" env:"DB_SSL_MODE" default:"disable"`
	MaxOpenConns    int           `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS" default:"25"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS" default:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"DB_CONN_MAX_LIFETIME" default:"5m"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env:"DB_CONN_MAX_IDLE_TIME" default:"1m"`
}

// DSN возвращает строку подключения к PostgreSQL
func (c *Config) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode)
}

// DefaultConfig возвращает конфигурацию по умолчанию
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "",
		Database:        "gophkeeper",
		SSLMode:         "disable",
		MaxOpenConns:    25,
		MaxIdleConns:    5,
		ConnMaxLifetime: 5 * time.Minute,
		ConnMaxIdleTime: time.Minute,
	}
}
