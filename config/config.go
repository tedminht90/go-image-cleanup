// config/config.go
package config

import (
	"go-image-cleanup/internal/infrastructure/logger"

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

func LoadConfig() (*Config, error) {
    viper.SetConfigFile(".env")
    viper.AutomaticEnv()

    // Set defaults
    viper.SetDefault("CLEANUP_SCHEDULE", "0 0 * * *")
    viper.SetDefault("HTTP_PORT", "8080")
    
    // Logger defaults
    viper.SetDefault("LOG_LEVEL", "info")
    viper.SetDefault("LOG_DIR", "/var/log/image-cleanup")
    viper.SetDefault("LOG_MAX_SIZE", 100)    // 100MB
    viper.SetDefault("LOG_MAX_BACKUPS", 5)   // 5 files
    viper.SetDefault("LOG_MAX_AGE", 30)      // 30 days
    viper.SetDefault("LOG_COMPRESS", true)

    if err := viper.ReadInConfig(); err != nil {
        return nil, err
    }

    return &Config{
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
    }, nil
}