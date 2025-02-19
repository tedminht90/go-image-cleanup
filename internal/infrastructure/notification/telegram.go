package notification

import (
	"fmt"
	"go-image-cleanup/internal/domain/notification"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

type TelegramNotifier struct {
	botToken string
	chatID   string
	logger   *zap.Logger
}

// Verify that TelegramNotifier implements Notifier interface
var _ notification.Notifier = (*TelegramNotifier)(nil)

func NewTelegramNotifier(botToken, chatID string, logger *zap.Logger) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		logger:   logger,
	}
}

func (n *TelegramNotifier) SendNotification(message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.botToken)

	resp, err := http.PostForm(apiURL, url.Values{
		"chat_id": {n.chatID},
		"text":    {message},
	})
	if err != nil {
		return fmt.Errorf("failed to send telegram message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram API returned non-OK status: %d", resp.StatusCode)
	}

	n.logger.Info("Successfully sent telegram notification",
		zap.String("chat_id", n.chatID))

	return nil
}
