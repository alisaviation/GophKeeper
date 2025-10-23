package config

import (
	"time"
)

// ClientConfig represents configuration for GophKeeper client
type ClientConfig struct {
	ServerAddress string
	StoragePath   string
	AutoSync      bool
}

// ServerConfig represents configuration for GophKeeper server
type ServerConfig struct {
	ServerAddress string
	GRPCPort      int
	Database      DatabaseConfig
	JWT           JWTConfig
	Encryption    EncryptionConfig
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Type            string
	Host            string
	Port            int
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

// EncryptionConfig represents encryption configuration
type EncryptionConfig struct {
	Key string
}

// FileConfig represents configuration file structure
type FileConfig struct {
	ServerAddress string `json:"server_address"`
	StoragePath   string `json:"storage_path"`
	AutoSync      bool   `json:"auto_sync"`

	GRPCPort int `json:"grpc_port"`

	DatabaseHost        string `json:"database_host"`
	DatabasePort        int    `json:"database_port"`
	DatabaseUser        string `json:"database_user"`
	DatabasePassword    string `json:"database_password"`
	DatabaseName        string `json:"database_name"`
	DatabaseSSLMode     string `json:"database_ssl_mode"`
	DatabaseMaxConns    int    `json:"database_max_conns"`
	DatabaseMaxIdleTime string `json:"database_max_idle_time"`

	JWTSecret        string `json:"jwt_secret"`
	JWTAccessExpiry  string `json:"jwt_access_expiry"`
	JWTRefreshExpiry string `json:"jwt_refresh_expiry"`

	EncryptionKey string `json:"encryption_key"`
}
