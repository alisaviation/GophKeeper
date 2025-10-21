// Package config handles configuration management for GophKeeper client and server.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
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

// FileConfig represents configuration file structure (JSON only)
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

// SetClientConfig initializes and returns client configuration
func SetClientConfig() ClientConfig {
	var config ClientConfig
	var configFile string

	flag.StringVar(&configFile, "c", "", "Path to config file")
	flag.StringVar(&configFile, "config", "", "Path to config file")
	serverAddress := flag.String("a", "localhost:8081", "Server address")
	storagePath := flag.String("s", "", "Storage path")
	autoSync := flag.Bool("auto-sync", true, "Enable auto sync")

	flag.Parse()

	// Default configuration
	defaultConfig := ClientConfig{
		ServerAddress: "localhost:8081",
		StoragePath:   getDefaultStoragePath(),
		AutoSync:      true,
	}

	config = defaultConfig

	// Load from config file if provided
	if configFile != "" {
		var fileConfig FileConfig
		if err := loadConfigFromFile(configFile, &fileConfig); err != nil {
			fmt.Printf("Warning: failed to load config file: %v\n", err)
		} else {
			applyFileConfigToClient(&config, fileConfig)
		}
	}

	// Apply command line flags
	config.ServerAddress = *serverAddress
	if *storagePath != "" {
		config.StoragePath = *storagePath
	}
	config.AutoSync = *autoSync

	// Apply environment variables
	applyEnvToClient(&config)

	// Apply config file from environment if not already loaded
	if envConfigFile, exists := os.LookupEnv("CONFIG"); exists && configFile == "" {
		var fileConfig FileConfig
		if err := loadConfigFromFile(envConfigFile, &fileConfig); err == nil {
			applyFileConfigToClient(&config, fileConfig)
		}
	}

	return config
}

// SetServerConfig initializes and returns server configuration
func SetServerConfig() ServerConfig {
	var config ServerConfig
	var configFile string

	flag.StringVar(&configFile, "c", "", "Path to config file")
	flag.StringVar(&configFile, "config", "", "Path to config file")
	grpcPort := flag.Int("p", 8080, "gRPC server port")

	// Database flags
	dbHost := flag.String("db-host", "localhost", "Database host")
	dbPort := flag.Int("db-port", 5432, "Database port")
	dbUser := flag.String("db-user", "postgres", "Database user")
	dbPassword := flag.String("db-password", "", "Database password")
	dbName := flag.String("db-name", "gophkeeper", "Database name")
	dbSSLMode := flag.String("db-ssl-mode", "disable", "Database SSL mode")
	dbMaxConns := flag.Int("db-max-conns", 25, "Database max connections")
	dbMaxIdleTime := flag.String("db-max-idle-time", "1m", "Database max idle time")

	// JWT flags
	jwtSecret := flag.String("jwt-secret", "default-jwt-secret-key-change-in-production", "JWT secret")
	jwtAccessExpiry := flag.String("jwt-access-expiry", "15m", "JWT access token expiry")
	jwtRefreshExpiry := flag.String("jwt-refresh-expiry", "168h", "JWT refresh token expiry")

	// Encryption flags
	encryptionKey := flag.String("encryption-key", "", "Encryption key")

	flag.Parse()

	// Default configuration
	defaultConfig := ServerConfig{
		GRPCPort: 8080,
		Database: DatabaseConfig{
			Type:            "postgres",
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "",
			Name:            "gophkeeper",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: time.Minute,
		},
		JWT: JWTConfig{
			Secret:        "default-jwt-secret-key-change-in-production",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
		Encryption: EncryptionConfig{
			Key: "",
		},
	}

	config = defaultConfig

	// Load from config file if provided
	if configFile != "" {
		var fileConfig FileConfig
		if err := loadConfigFromFile(configFile, &fileConfig); err != nil {
			fmt.Printf("Warning: failed to load config file: %v\n", err)
		} else {
			applyFileConfigToServer(&config, fileConfig)
		}
	}

	// Apply command line flags
	config.GRPCPort = *grpcPort
	config.Database.Host = *dbHost
	config.Database.Port = *dbPort
	config.Database.User = *dbUser
	config.Database.Password = *dbPassword
	config.Database.Name = *dbName
	config.Database.SSLMode = *dbSSLMode
	config.Database.MaxOpenConns = *dbMaxConns
	if *dbMaxIdleTime != "" {
		config.Database.ConnMaxIdleTime = parseDuration(*dbMaxIdleTime, config.Database.ConnMaxIdleTime)
	}

	config.JWT.Secret = *jwtSecret
	if *jwtAccessExpiry != "" {
		config.JWT.AccessExpiry = parseDuration(*jwtAccessExpiry, config.JWT.AccessExpiry)
	}
	if *jwtRefreshExpiry != "" {
		config.JWT.RefreshExpiry = parseDuration(*jwtRefreshExpiry, config.JWT.RefreshExpiry)
	}

	config.Encryption.Key = *encryptionKey

	// Apply environment variables
	applyEnvToServer(&config)

	// Apply config file from environment if not already loaded
	if envConfigFile, exists := os.LookupEnv("CONFIG"); exists && configFile == "" {
		var fileConfig FileConfig
		if err := loadConfigFromFile(envConfigFile, &fileConfig); err == nil {
			applyFileConfigToServer(&config, fileConfig)
		}
	}

	return config
}

// Helper functions

func getDefaultStoragePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "./gophkeeper"
	}
	return configDir + "/gophkeeper"
}

func parseDuration(value string, defaultValue time.Duration) time.Duration {
	if value == "" {
		return defaultValue
	}
	duration, err := time.ParseDuration(value)
	if err != nil {
		fmt.Printf("Warning: invalid duration '%s', using default: %v\n", value, err)
		return defaultValue
	}
	return duration
}

func applyFileConfigToClient(config *ClientConfig, fileConfig FileConfig) {
	if fileConfig.ServerAddress != "" {
		config.ServerAddress = fileConfig.ServerAddress
	}
	if fileConfig.StoragePath != "" {
		config.StoragePath = fileConfig.StoragePath
	}
	// AutoSync doesn't have a "not set" state, so we always apply it from file if present
	config.AutoSync = fileConfig.AutoSync
}

func applyFileConfigToServer(config *ServerConfig, fileConfig FileConfig) {
	if fileConfig.GRPCPort != 0 {
		config.GRPCPort = fileConfig.GRPCPort
	}
	if fileConfig.ServerAddress != "" {
		config.ServerAddress = fileConfig.ServerAddress
	}

	// Database configuration
	if fileConfig.DatabaseHost != "" {
		config.Database.Host = fileConfig.DatabaseHost
	}
	if fileConfig.DatabasePort != 0 {
		config.Database.Port = fileConfig.DatabasePort
	}
	if fileConfig.DatabaseUser != "" {
		config.Database.User = fileConfig.DatabaseUser
	}
	if fileConfig.DatabasePassword != "" {
		config.Database.Password = fileConfig.DatabasePassword
	}
	if fileConfig.DatabaseName != "" {
		config.Database.Name = fileConfig.DatabaseName

	}
	if fileConfig.DatabaseSSLMode != "" {
		config.Database.SSLMode = fileConfig.DatabaseSSLMode

	}
	if fileConfig.DatabaseMaxConns != 0 {
		config.Database.MaxOpenConns = fileConfig.DatabaseMaxConns

	}
	if fileConfig.DatabaseMaxIdleTime != "" {
		config.Database.ConnMaxIdleTime = parseDuration(fileConfig.DatabaseMaxIdleTime, config.Database.ConnMaxIdleTime)
	}

	// JWT configuration
	if fileConfig.JWTSecret != "" {
		config.JWT.Secret = fileConfig.JWTSecret

	}
	if fileConfig.JWTAccessExpiry != "" {
		config.JWT.AccessExpiry = parseDuration(fileConfig.JWTAccessExpiry, config.JWT.AccessExpiry)
	}
	if fileConfig.JWTRefreshExpiry != "" {
		config.JWT.RefreshExpiry = parseDuration(fileConfig.JWTRefreshExpiry, config.JWT.RefreshExpiry)
	}

	// Encryption configuration
	if fileConfig.EncryptionKey != "" {
		config.Encryption.Key = fileConfig.EncryptionKey
	}
}

func applyEnvToClient(config *ClientConfig) {
	if envServerAddress, exists := os.LookupEnv("SERVER_ADDRESS"); exists {
		config.ServerAddress = envServerAddress
	}
	if envStoragePath, exists := os.LookupEnv("STORAGE_PATH"); exists {
		config.StoragePath = envStoragePath
	}
	if envAutoSync, exists := os.LookupEnv("AUTO_SYNC"); exists {
		if autoSync, err := strconv.ParseBool(envAutoSync); err == nil {
			config.AutoSync = autoSync
		}
	}
}

func applyEnvToServer(config *ServerConfig) {
	if envGRPCPort, exists := os.LookupEnv("GRPC_PORT"); exists {
		if port, err := strconv.Atoi(envGRPCPort); err == nil {
			config.GRPCPort = port
		}
	}
	if envServerAddress, exists := os.LookupEnv("SERVER_ADDRESS"); exists {
		config.ServerAddress = envServerAddress
	}

	// Database environment variables
	if envDBHost, exists := os.LookupEnv("DB_HOST"); exists {
		config.Database.Host = envDBHost
	}
	if envDBPort, exists := os.LookupEnv("DB_PORT"); exists {
		if port, err := strconv.Atoi(envDBPort); err == nil {
			config.Database.Port = port
		}
	}
	if envDBUser, exists := os.LookupEnv("DB_USER"); exists {
		config.Database.User = envDBUser
	}
	if envDBPassword, exists := os.LookupEnv("DB_PASSWORD"); exists {
		config.Database.Password = envDBPassword
	}
	if envDBName, exists := os.LookupEnv("DB_NAME"); exists {
		config.Database.Name = envDBName
	}
	if envDBSSLMode, exists := os.LookupEnv("DB_SSL_MODE"); exists {
		config.Database.SSLMode = envDBSSLMode
	}
	if envDBMaxConns, exists := os.LookupEnv("DB_MAX_CONNS"); exists {
		if maxConns, err := strconv.Atoi(envDBMaxConns); err == nil {
			config.Database.MaxOpenConns = maxConns
		}
	}
	if envDBMaxIdleTime, exists := os.LookupEnv("DB_MAX_IDLE_TIME"); exists {
		config.Database.ConnMaxIdleTime = parseDuration(envDBMaxIdleTime, config.Database.ConnMaxIdleTime)
	}

	// JWT environment variables
	if envJWTSecret, exists := os.LookupEnv("JWT_SECRET"); exists {
		config.JWT.Secret = envJWTSecret
	}
	if envJWTAccessExpiry, exists := os.LookupEnv("JWT_ACCESS_EXPIRY"); exists {
		config.JWT.AccessExpiry = parseDuration(envJWTAccessExpiry, config.JWT.AccessExpiry)
	}
	if envJWTRefreshExpiry, exists := os.LookupEnv("JWT_REFRESH_EXPIRY"); exists {
		config.JWT.RefreshExpiry = parseDuration(envJWTRefreshExpiry, config.JWT.RefreshExpiry)
	}

	// Encryption environment variables
	if envEncryptionKey, exists := os.LookupEnv("ENCRYPTION_KEY"); exists {
		config.Encryption.Key = envEncryptionKey
	}
}

// loadConfigFromFile loads configuration from JSON file only
func loadConfigFromFile(filePath string, config *FileConfig) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, config); err != nil {
		return fmt.Errorf("failed to parse JSON config: %w", err)
	}

	return nil
}

// DefaultClientConfig returns default client configuration
func DefaultClientConfig() ClientConfig {
	return ClientConfig{
		ServerAddress: "localhost:8090",
		StoragePath:   getDefaultStoragePath(),
		AutoSync:      true,
	}
}

// DefaultServerConfig returns default server configuration
func DefaultServerConfig() ServerConfig {
	return ServerConfig{
		ServerAddress: "localhost:8080",
		GRPCPort:      8080,
		Database: DatabaseConfig{
			Type:            "postgres",
			Host:            "localhost",
			Port:            5432,
			User:            "postgres",
			Password:        "",
			Name:            "gophkeeper",
			SSLMode:         "disable",
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: time.Minute,
		},
		JWT: JWTConfig{
			Secret:        "default-jwt-secret-key-change-in-production",
			AccessExpiry:  15 * time.Minute,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
		Encryption: EncryptionConfig{
			Key: "",
		},
	}
}

// DSN возвращает строку подключения к базе данных
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.Name, d.SSLMode)
}
