package messaging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// TelegramConfig نود Telegram
type TelegramConfig struct {
	BotToken string `json:"botToken"` // Bot token
	ChatID   string `json:"chatId"`   // Chat ID (optional)
	Mode     string `json:"mode"`     // send or receive
}

// TelegramExecutor اجراکننده نود Telegram
type TelegramExecutor struct {
	config     TelegramConfig
	client     *http.Client
	offset     int64
	outputChan chan node.Message
	stopChan   chan struct{}
}

// Init initializes the executor with configuration
func (e *TelegramExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var telegramConfig TelegramConfig
	if err := json.Unmarshal(configJSON, &telegramConfig); err != nil {
		return fmt.Errorf("invalid telegram config: %w", err)
	}

	if telegramConfig.BotToken == "" {
		return fmt.Errorf("bot token is required")
	}

	if telegramConfig.Mode == "" {
		telegramConfig.Mode = "send"
	}

	e.config = telegramConfig
	e.client = &http.Client{Timeout: 30 * time.Second}
	e.outputChan = make(chan node.Message, 10)
	e.stopChan = make(chan struct{})

	return nil
}

// Execute اجرای نود
func (e *TelegramExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if e.config.Mode == "send" {
		return e.sendMessage(ctx, msg)
	} else {
		// Start polling if not started
		if e.offset == 0 {
			go e.pollUpdates()
		}

		// Wait for incoming message
		select {
		case <-ctx.Done():
			return node.Message{}, ctx.Err()
		case telegramMsg := <-e.outputChan:
			return telegramMsg, nil
		}
	}
}

// sendMessage ارسال پیام
func (e *TelegramExecutor) sendMessage(ctx context.Context, msg node.Message) (node.Message, error) {
	var text string
	var chatID string
	var parseMode string

	// Get message content
	if t, ok := msg.Payload["text"].(string); ok {
		text = t
	} else if t, ok := msg.Payload["message"].(string); ok {
		text = t
	} else if t, ok := msg.Payload["payload"].(string); ok {
		text = t
	} else {
		// Convert payload to JSON string
		jsonData, _ := json.Marshal(msg.Payload)
		text = string(jsonData)
	}

	if c, ok := msg.Payload["chatId"].(string); ok {
		chatID = c
	}
	if p, ok := msg.Payload["parseMode"].(string); ok {
		parseMode = p
	}

	if text == "" {
		return node.Message{}, fmt.Errorf("no text to send")
	}

	// Use config chatID if not provided in message
	if chatID == "" {
		chatID = e.config.ChatID
	}
	if chatID == "" {
		return node.Message{}, fmt.Errorf("chat ID is required")
	}

	// Prepare request
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", e.config.BotToken)
	payload := map[string]interface{}{
		"chat_id": chatID,
		"text":    text,
	}
	if parseMode != "" {
		payload["parse_mode"] = parseMode
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return node.Message{}, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return node.Message{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(req)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to send message: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if resp.StatusCode != 200 {
		return node.Message{}, fmt.Errorf("telegram API error: %s", string(body))
	}

	return node.Message{
		Payload: map[string]interface{}{
			"sent":   true,
			"chatId": chatID,
			"result": result,
		},
	}, nil
}

// pollUpdates دریافت پیام‌ها
func (e *TelegramExecutor) pollUpdates() {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates", e.config.BotToken)

	for {
		select {
		case <-e.stopChan:
			return
		default:
		}

		// Prepare request
		payload := map[string]interface{}{
			"offset":  e.offset,
			"timeout": 30,
		}
		jsonData, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := e.client.Do(req)
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		var result struct {
			Ok     bool `json:"ok"`
			Result []struct {
				UpdateID int64 `json:"update_id"`
				Message  struct {
					MessageID int64  `json:"message_id"`
					From      struct {
						ID        int64  `json:"id"`
						FirstName string `json:"first_name"`
						Username  string `json:"username"`
					} `json:"from"`
					Chat struct {
						ID int64 `json:"id"`
					} `json:"chat"`
					Date int64  `json:"date"`
					Text string `json:"text"`
				} `json:"message"`
			} `json:"result"`
		}

		if err := json.Unmarshal(body, &result); err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if !result.Ok {
			time.Sleep(5 * time.Second)
			continue
		}

		// Process updates
		for _, update := range result.Result {
			e.offset = update.UpdateID + 1

			msg := node.Message{
				Payload: map[string]interface{}{
					"updateId":  update.UpdateID,
					"messageId": update.Message.MessageID,
					"from": map[string]interface{}{
						"id":        update.Message.From.ID,
						"firstName": update.Message.From.FirstName,
						"username":  update.Message.From.Username,
					},
					"chatId": update.Message.Chat.ID,
					"date":   update.Message.Date,
					"text":   update.Message.Text,
				},
			}

			select {
			case e.outputChan <- msg:
			default:
			}
		}
	}
}

// Cleanup پاکسازی منابع
func (e *TelegramExecutor) Cleanup() error {
	close(e.stopChan)
	close(e.outputChan)
	return nil
}
