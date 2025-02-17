// config/config.go
package config

import (
	"fmt"
	"go-image-cleanup/internal/infrastructure/logger"
	"go-image-cleanup/pkg/helper"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken string
	TelegramChatID   string
	CleanupSchedule  string
	HTTPPort         string

	// Logger config
	Logger logger.Config
}

// String returns a formatted string of configuration (excluding sensitive data)
func (c *Config) String() string {
	var sb strings.Builder
	sb.WriteString("\nConfiguration:\n")
	sb.WriteString("----------------\n")
	// Hide sensitive information
	sb.WriteString(fmt.Sprintf("TELEGRAM_BOT_TOKEN: %s\n", helper.MaskValue(c.TelegramBotToken)))
	sb.WriteString(fmt.Sprintf("TELEGRAM_CHAT_ID: %s\n", helper.MaskValue(c.TelegramChatID)))
	sb.WriteString(fmt.Sprintf("CLEANUP_SCHEDULE: %s\n", c.CleanupSchedule))
	sb.WriteString(fmt.Sprintf("HTTP_PORT: %s\n", c.HTTPPort))
	sb.WriteString("\nLogger Configuration:\n")
	sb.WriteString("--------------------\n")
	sb.WriteString(fmt.Sprintf("LOG_LEVEL: %s\n", c.Logger.Level))
	sb.WriteString(fmt.Sprintf("LOG_DIR: %s\n", c.Logger.LogDir))
	sb.WriteString(fmt.Sprintf("LOG_MAX_SIZE: %d MB\n", c.Logger.MaxSize))
	sb.WriteString(fmt.Sprintf("LOG_MAX_BACKUPS: %d files\n", c.Logger.MaxBackups))
	sb.WriteString(fmt.Sprintf("LOG_MAX_AGE: %d days\n", c.Logger.MaxAge))
	sb.WriteString(fmt.Sprintf("LOG_COMPRESS: %v\n", c.Logger.Compress))
	return sb.String()
}

func LoadConfig() (*Config, error) {
	// Set default config path
	viper.SetConfigFile("/etc/image-cleanup/.env")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("CLEANUP_SCHEDULE", "0 0 * * *")
	viper.SetDefault("HTTP_PORT", "8080")

	// Logger defaults
	viper.SetDefault("LOG_LEVEL", "info")
	viper.SetDefault("LOG_DIR", "/var/log/image-cleanup")
	viper.SetDefault("LOG_MAX_SIZE", 100)  // 100MB
	viper.SetDefault("LOG_MAX_BACKUPS", 5) // 5 files
	viper.SetDefault("LOG_MAX_AGE", 30)    // 30 days
	viper.SetDefault("LOG_COMPRESS", true)

	// Read config file
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Printf("Error reading config file: %v\n", err)
		fmt.Println("Using environment variables and defaults")
	} else {
		fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
	}

	// Create config structure
	config := &Config{
		TelegramBotToken: viper.GetString("TELEGRAM_BOT_TOKEN"),
		TelegramChatID:   viper.GetString("TELEGRAM_CHAT_ID"),
		CleanupSchedule:  viper.GetString("CLEANUP_SCHEDULE"),
		HTTPPort:         viper.GetString("HTTP_PORT"),
		Logger: logger.Config{
			Level:      viper.GetString("LOG_LEVEL"),
			LogDir:     viper.GetString("LOG_DIR"),
			MaxSize:    viper.GetInt("LOG_MAX_SIZE"),
			MaxBackups: viper.GetInt("LOG_MAX_BACKUPS"),
			MaxAge:     viper.GetInt("LOG_MAX_AGE"),
			Compress:   viper.GetBool("LOG_COMPRESS"),
		},
	}

	// Print current configuration
	fmt.Println(config.String())

	return config, nil
}
