package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/edgeflow/edgeflow/internal/node"
)

// EmailConfig نود Email
type EmailConfig struct {
	Host     string `json:"host"`     // SMTP host
	Port     int    `json:"port"`     // SMTP port
	Username string `json:"username"` // Username
	Password string `json:"password"` // Password
	From     string `json:"from"`     // From address
	To       string `json:"to"`       // Default to address
	UseTLS   bool   `json:"useTls"`   // Use TLS
}

// EmailExecutor اجراکننده نود Email
type EmailExecutor struct {
	config EmailConfig
}

// Init initializes the executor with configuration
func (e *EmailExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var emailConfig EmailConfig
	if err := json.Unmarshal(configJSON, &emailConfig); err != nil {
		return fmt.Errorf("invalid email config: %w", err)
	}

	// Validate
	if emailConfig.Host == "" {
		return fmt.Errorf("host is required")
	}
	if emailConfig.Port == 0 {
		emailConfig.Port = 587
	}

	e.config = emailConfig
	return nil
}

// Execute اجرای نود
func (e *EmailExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	var to, subject, body string

	// Get email content from message
	if t, ok := msg.Payload["to"].(string); ok {
		to = t
	}
	if s, ok := msg.Payload["subject"].(string); ok {
		subject = s
	}
	if b, ok := msg.Payload["body"].(string); ok {
		body = b
	} else if b, ok := msg.Payload["text"].(string); ok {
		body = b
	} else if b, ok := msg.Payload["message"].(string); ok {
		body = b
	}

	// Use config defaults if not provided
	if to == "" {
		to = e.config.To
	}
	if to == "" {
		return node.Message{}, fmt.Errorf("to address is required")
	}
	if body == "" {
		return node.Message{}, fmt.Errorf("email body is required")
	}
	if subject == "" {
		subject = "EdgeFlow Notification"
	}

	// Prepare email
	from := e.config.From
	if from == "" {
		from = e.config.Username
	}

	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", from, to, subject, body)

	// Send email
	auth := smtp.PlainAuth("", e.config.Username, e.config.Password, e.config.Host)
	addr := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)

	toList := strings.Split(to, ",")
	for i := range toList {
		toList[i] = strings.TrimSpace(toList[i])
	}

	err := smtp.SendMail(addr, auth, from, toList, []byte(message))
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to send email: %w", err)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"sent":    true,
			"to":      to,
			"subject": subject,
			"from":    from,
		},
	}, nil
}

// Cleanup پاکسازی منابع
func (e *EmailExecutor) Cleanup() error {
	return nil
}
