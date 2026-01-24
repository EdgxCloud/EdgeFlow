package dashboard

import (
	"context"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// NotificationType defines the notification type
type NotificationType string

const (
	NotificationTypeSuccess NotificationType = "success"
	NotificationTypeInfo    NotificationType = "info"
	NotificationTypeWarning NotificationType = "warning"
	NotificationTypeError   NotificationType = "error"
)

// NotificationNode displays notifications/toasts on the dashboard
type NotificationNode struct {
	*BaseWidget
	notifType NotificationType
	duration  int // Duration in milliseconds
	position  string
	closable  bool
}

// NewNotificationNode creates a new notification widget
func NewNotificationNode() *NotificationNode {
	return &NotificationNode{
		BaseWidget: NewBaseWidget(WidgetTypeNotification),
		notifType:  NotificationTypeInfo,
		duration:   3000,
		position:   "top-right",
		closable:   true,
	}
}

// Init initializes the notification node
func (n *NotificationNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	if notifType, ok := config["type"].(string); ok {
		n.notifType = NotificationType(notifType)
	}
	if duration, ok := config["duration"].(float64); ok {
		n.duration = int(duration)
	}
	if position, ok := config["position"].(string); ok {
		n.position = position
	}
	if closable, ok := config["closable"].(bool); ok {
		n.closable = closable
	}

	return nil
}

// Execute handles notification display
func (n *NotificationNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract notification message
	var message string
	var title string

	if msgVal, ok := msg.Payload["message"].(string); ok {
		message = msgVal
	} else if msgVal, ok := msg.Payload["value"].(string); ok {
		message = msgVal
	}

	if titleVal, ok := msg.Payload["title"].(string); ok {
		title = titleVal
	}

	// Override type if provided
	notifType := n.notifType
	if typeVal, ok := msg.Payload["type"].(string); ok {
		notifType = NotificationType(typeVal)
	}

	// Update dashboard with notification
	if n.manager != nil {
		notificationData := map[string]interface{}{
			"title":     title,
			"message":   message,
			"type":      notifType,
			"duration":  n.duration,
			"position":  n.position,
			"closable":  n.closable,
			"timestamp": time.Now().Unix(),
		}
		n.manager.UpdateWidget(n.id, notificationData)
	}

	return node.Message{}, nil
}

// SetManager sets the dashboard manager
func (n *NotificationNode) SetManager(manager *Manager) {
	n.manager = manager
}
