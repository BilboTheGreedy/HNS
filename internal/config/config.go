package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	DNS      DNSConfig
	Logging  LoggingConfig
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Host          string
	Port          int
	User          string
	Password      string
	DBName        string
	SSLMode       string
	PoolSize      int
	RunMigrations bool
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret        string
	JWTExpiration    time.Duration
	APIKeyExpiration time.Duration
}

// DNSConfig holds DNS configuration
type DNSConfig struct {
	Servers []string
	Timeout time.Duration
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string
	Format string
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// Default values
	setDefaults()

	// Enable environment variable overrides
	viper.AutomaticEnv()

	configErr := viper.ReadInConfig()
	if configErr != nil {
		if _, ok := configErr.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, continue with defaults and env vars
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %v", configErr)
		}
	} else {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found, continue with defaults and env vars
			fmt.Println("Config file not found, using defaults and environment variables")
		} else {
			return nil, fmt.Errorf("error reading config file: %v", err)
		}
	}

	config := &Config{
		Server: ServerConfig{
			Port:            viper.GetInt("server.port"),
			ReadTimeout:     viper.GetDuration("server.readTimeout"),
			WriteTimeout:    viper.GetDuration("server.writeTimeout"),
			ShutdownTimeout: viper.GetDuration("server.shutdownTimeout"),
		},
		Database: DatabaseConfig{
			Host:          viper.GetString("database.host"),
			Port:          viper.GetInt("database.port"),
			User:          viper.GetString("database.user"),
			Password:      viper.GetString("database.password"),
			DBName:        viper.GetString("database.dbname"),
			SSLMode:       viper.GetString("database.sslmode"),
			PoolSize:      viper.GetInt("database.poolSize"),
			RunMigrations: viper.GetBool("database.runMigrations"),
		},
		Auth: AuthConfig{
			JWTSecret:        viper.GetString("auth.jwtSecret"),
			JWTExpiration:    viper.GetDuration("auth.jwtExpiration"),
			APIKeyExpiration: viper.GetDuration("auth.apiKeyExpiration"),
		},
		DNS: DNSConfig{
			Servers: viper.GetStringSlice("dns.servers"),
			Timeout: viper.GetDuration("dns.timeout"),
		},
		Logging: LoggingConfig{
			Level:  viper.GetString("logging.level"),
			Format: viper.GetString("logging.format"),
		},
	}

	return config, nil
}

// setDefaults sets default values for configuration
func setDefaults() {
	// Server defaults
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.readTimeout", "15s")
	viper.SetDefault("server.writeTimeout", "15s")
	viper.SetDefault("server.shutdownTimeout", "5s")

	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "postgres")
	viper.SetDefault("database.dbname", "hns")
	viper.SetDefault("database.sslmode", "disable")
	viper.SetDefault("database.poolSize", 10)
	viper.SetDefault("database.runMigrations", true)

	// Auth defaults
	viper.SetDefault("auth.jwtSecret", "supersecretkey")
	viper.SetDefault("auth.jwtExpiration", "24h")
	viper.SetDefault("auth.apiKeyExpiration", "720h") // 30 days

	// DNS defaults
	viper.SetDefault("dns.servers", []string{"8.8.8.8", "8.8.4.4"})
	viper.SetDefault("dns.timeout", "5s")

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
}
