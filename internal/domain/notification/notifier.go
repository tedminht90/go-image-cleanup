package notification

type Notifier interface {
    SendNotification(message string) error
}