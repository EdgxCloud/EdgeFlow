package network

import (
	"context"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

type XMLParserNode struct {
	action string // "parse" or "stringify"
}

func NewXMLParserNode() *XMLParserNode {
	return &XMLParserNode{
		action: "parse",
	}
}

func (n *XMLParserNode) Init(config map[string]interface{}) error {
	if action, ok := config["action"].(string); ok {
		n.action = action
	}
	return nil
}

func (n *XMLParserNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	switch n.action {
	case "parse":
		return n.parseXML(msg)
	case "stringify":
		return n.stringifyXML(msg)
	default:
		return msg, fmt.Errorf("unknown action: %s", n.action)
	}
}

func (n *XMLParserNode) parseXML(msg node.Message) (node.Message, error) {
	var xmlStr string

	// Handle data from payload
	if str, ok := msg.Payload["data"].(string); ok {
		xmlStr = str
	} else if bytes, ok := msg.Payload["data"].([]byte); ok {
		xmlStr = string(bytes)
	} else {
		return msg, fmt.Errorf("payload must contain 'data' as string or []byte")
	}

	var result map[string]interface{}
	decoder := xml.NewDecoder(strings.NewReader(xmlStr))

	var current map[string]interface{}
	var stack []map[string]interface{}

	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			elem := make(map[string]interface{})
			for _, attr := range t.Attr {
				elem["@"+attr.Name.Local] = attr.Value
			}
			if current != nil {
				stack = append(stack, current)
			}
			current = elem

		case xml.CharData:
			if current != nil {
				text := strings.TrimSpace(string(t))
				if text != "" {
					current["#text"] = text
				}
			}

		case xml.EndElement:
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				parent[t.Name.Local] = current
				current = parent
			} else {
				result = current
			}
		}
	}

	msg.Payload = map[string]interface{}{"data": result}
	return msg, nil
}

func (n *XMLParserNode) stringifyXML(msg node.Message) (node.Message, error) {
	var dataToStringify interface{}

	// Extract data from payload
	if data, ok := msg.Payload["data"]; ok {
		dataToStringify = data
	} else {
		dataToStringify = msg.Payload
	}

	data, err := xml.Marshal(dataToStringify)
	if err != nil {
		return msg, fmt.Errorf("failed to stringify XML: %w", err)
	}

	msg.Payload = map[string]interface{}{"data": string(data)}
	return msg, nil
}

func (n *XMLParserNode) Cleanup() error {
	return nil
}

func NewXMLParserExecutor() node.Executor {
	return NewXMLParserNode()
}
