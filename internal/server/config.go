package server

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/viper"
)

// Config holds the MCP server configuration
type Config struct {
	// API Key authentication (simpler, read-only endpoints)
	APIKey string

	// OAuth2 authentication (full API access)
	ClientID     string
	ClientSecret string

	// Optional settings
	LogLevel     slog.Level
	CacheEnabled bool
	CacheTTL     int // hours
	DefaultLang  string
}

// LoadConfig loads configuration from environment, file, and flags
// Priority: Environment > Config File > Defaults
func LoadConfig(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("cache_enabled", true)
	v.SetDefault("cache_ttl_hours", 24)
	v.SetDefault("default_language", "en")
	v.SetDefault("log_level", "info")

	// Environment variables (highest priority)
	v.SetEnvPrefix("OPENPLANTBOOK")
	v.AutomaticEnv()

	// Config file (if provided)
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			// Only fail if file was explicitly provided but couldn't be read
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return nil, fmt.Errorf("read config file: %w", err)
			}
		}
	} else {
		// Try default config locations
		home, err := os.UserHomeDir()
		if err == nil {
			v.AddConfigPath(home + "/.config/openplantbook-mcp")
			v.AddConfigPath(home)
			v.SetConfigName("config")
			v.SetConfigType("json")
			// Ignore errors for optional config file
			_ = v.ReadInConfig()
		}
	}

	// Parse and validate
	config := &Config{
		APIKey:       v.GetString("api_key"),
		ClientID:     v.GetString("client_id"),
		ClientSecret: v.GetString("client_secret"),
		CacheEnabled: v.GetBool("cache_enabled"),
		CacheTTL:     v.GetInt("cache_ttl_hours"),
		DefaultLang:  v.GetString("default_language"),
	}

	// Parse log level
	switch v.GetString("log_level") {
	case "debug":
		config.LogLevel = slog.LevelDebug
	case "info":
		config.LogLevel = slog.LevelInfo
	case "warn":
		config.LogLevel = slog.LevelWarn
	case "error":
		config.LogLevel = slog.LevelError
	default:
		config.LogLevel = slog.LevelInfo
	}

	// Validate: need EITHER API key OR OAuth2 credentials
	hasAPIKey := config.APIKey != ""
	hasOAuth2 := config.ClientID != "" && config.ClientSecret != ""

	if !hasAPIKey && !hasOAuth2 {
		return nil, fmt.Errorf("authentication required: provide either api_key OR (client_id and client_secret)")
	}

	if hasAPIKey && hasOAuth2 {
		return nil, fmt.Errorf("multiple authentication methods provided: use either api_key OR OAuth2, not both")
	}

	return config, nil
}
