package network

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterAllNodes ثبت تمام نودهای network
func RegisterAllNodes(registry *node.Registry) {
	// HTTP Nodes
	registry.Register(&node.NodeInfo{
		Type:        "http-request",
		Name:        "HTTP Request",
		Category:    node.NodeTypeProcessing,
		Description: "ارسال درخواست HTTP با پشتیبانی از تمام متدها",
		Icon:        "globe",
		Factory:     NewHTTPRequestExecutor,
	})

	registry.Register(&node.NodeInfo{
		Type:        "http-webhook",
		Name:        "HTTP Webhook",
		Category:    node.NodeTypeInput,
		Description: "دریافت درخواست HTTP از طریق webhook",
		Icon:        "webhook",
		Factory:     NewHTTPWebhookExecutor,
	})

	registry.Register(&node.NodeInfo{
		Type:        "http-response",
		Name:        "HTTP Response",
		Category:    node.NodeTypeOutput,
		Description: "ارسال پاسخ HTTP",
		Icon:        "send",
		Factory:     NewHTTPResponseExecutor,
	})

	// MQTT Nodes
	registry.Register(&node.NodeInfo{
		Type:        "mqtt-in",
		Name:        "MQTT Input",
		Category:    node.NodeTypeInput,
		Description: "اشتراک در topic و دریافت پیام از MQTT broker",
		Icon:        "message-square",
		Factory:     NewMQTTInExecutor,
	})

	registry.Register(&node.NodeInfo{
		Type:        "mqtt-out",
		Name:        "MQTT Output",
		Category:    node.NodeTypeOutput,
		Description: "انتشار پیام به MQTT broker",
		Icon:        "send",
		Factory:     NewMQTTOutExecutor,
	})

	// WebSocket Node
	registry.Register(&node.NodeInfo{
		Type:        "websocket-client",
		Name:        "WebSocket Client",
		Category:    node.NodeTypeProcessing,
		Description: "اتصال به WebSocket server و ارسال/دریافت پیام",
		Icon:        "radio",
		Factory:     NewWebSocketClientExecutor,
	})

	// TCP Node
	registry.Register(&node.NodeInfo{
		Type:        "tcp-client",
		Name:        "TCP Client",
		Category:    node.NodeTypeProcessing,
		Description: "اتصال به TCP server و ارسال/دریافت داده",
		Icon:        "server",
		Factory:     NewTCPClientExecutor,
	})

	// UDP Node
	registry.Register(&node.NodeInfo{
		Type:        "udp",
		Name:        "UDP",
		Category:    node.NodeTypeProcessing,
		Description: "ارسال و دریافت داده از طریق UDP",
		Icon:        "radio",
		Factory:     NewUDPExecutor,
	})

	// Parser Nodes
	registry.Register(&node.NodeInfo{
		Type:        "json-parser",
		Name:        "JSON Parser",
		Category:    node.NodeTypeProcessing,
		Description: "تبدیل JSON به شیء و بالعکس",
		Icon:        "braces",
		Factory:     NewJSONParserExecutor,
	})

	registry.Register(&node.NodeInfo{
		Type:        "xml-parser",
		Name:        "XML Parser",
		Category:    node.NodeTypeProcessing,
		Description: "تبدیل XML به JSON و بالعکس",
		Icon:        "code",
		Factory:     NewXMLParserExecutor,
	})

	registry.Register(&node.NodeInfo{
		Type:        "csv-parser",
		Name:        "CSV Parser",
		Category:    node.NodeTypeProcessing,
		Description: "تبدیل CSV به JSON و بالعکس",
		Icon:        "table",
		Factory:     NewCSVParserExecutor,
	})

	registry.Register(&node.NodeInfo{
		Type:        "yaml-parser",
		Name:        "YAML Parser",
		Category:    node.NodeTypeProcessing,
		Description: "تبدیل YAML به JSON و بالعکس",
		Icon:        "file-text",
		Factory:     NewYAMLParserExecutor,
	})

	// File Nodes
	registry.Register(&node.NodeInfo{
		Type:        "watch",
		Name:        "File Watch",
		Category:    node.NodeTypeInput,
		Description: "نظارت بر تغییرات فایل/پوشه",
		Icon:        "eye",
		Factory:     NewWatchExecutor,
	})

	registry.Register(&node.NodeInfo{
		Type:        "file-in",
		Name:        "File Read",
		Category:    node.NodeTypeInput,
		Description: "خواندن محتوای فایل",
		Icon:        "file",
		Factory:     NewFileInExecutor,
	})

	registry.Register(&node.NodeInfo{
		Type:        "file-out",
		Name:        "File Write",
		Category:    node.NodeTypeOutput,
		Description: "نوشتن محتوا در فایل",
		Icon:        "save",
		Factory:     NewFileOutExecutor,
	})

	// Serial Nodes
	registry.Register(&node.NodeInfo{
		Type:        "serial-in",
		Name:        "Serial Input",
		Category:    node.NodeTypeInput,
		Description: "دریافت داده از پورت سریال",
		Icon:        "terminal",
		Factory:     NewSerialInExecutor,
	})

	registry.Register(&node.NodeInfo{
		Type:        "serial-out",
		Name:        "Serial Output",
		Category:    node.NodeTypeOutput,
		Description: "ارسال داده به پورت سریال",
		Icon:        "terminal",
		Factory:     NewSerialOutExecutor,
	})
}
