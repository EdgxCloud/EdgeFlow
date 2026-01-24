package network

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

type HTTPResponseNode struct {
	statusCode int
	headers    map[string]string
}

func NewHTTPResponseNode() *HTTPResponseNode {
	return &HTTPResponseNode{
		statusCode: 200,
		headers:    make(map[string]string),
	}
}

func (n *HTTPResponseNode) Init(config map[string]interface{}) error {
	if statusCode, ok := config["statusCode"].(float64); ok {
		n.statusCode = int(statusCode)
	} else if statusCode, ok := config["statusCode"].(int); ok {
		n.statusCode = statusCode
	}

	if headers, ok := config["headers"].(map[string]interface{}); ok {
		n.headers = make(map[string]string)
		for k, v := range headers {
			n.headers[k] = fmt.Sprintf("%v", v)
		}
	}

	return nil
}

func (n *HTTPResponseNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	response := map[string]interface{}{
		"statusCode": n.statusCode,
		"headers":    n.headers,
		"body":       msg.Payload,
	}

	payload := msg.Payload
	if statusCode, ok := payload["statusCode"]; ok {
		if sc, ok := statusCode.(float64); ok {
			response["statusCode"] = int(sc)
		} else if sc, ok := statusCode.(int); ok {
			response["statusCode"] = sc
		}
	}

	if headers, ok := payload["headers"]; ok {
		if h, ok := headers.(map[string]interface{}); ok {
			responseHeaders := make(map[string]string)
			for k, v := range n.headers {
				responseHeaders[k] = v
			}
			for k, v := range h {
				responseHeaders[k] = fmt.Sprintf("%v", v)
			}
			response["headers"] = responseHeaders
		}
	}

	if body, ok := payload["body"]; ok {
		response["body"] = body
	}

	msg.Payload = response
	msg.Topic = "http-response"

	return msg, nil
}

func (n *HTTPResponseNode) Cleanup() error {
	return nil
}

func NewHTTPResponseExecutor() node.Executor {
	return NewHTTPResponseNode()
}
