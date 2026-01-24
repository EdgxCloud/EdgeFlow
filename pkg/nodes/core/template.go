package core

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/edgeflow/edgeflow/internal/node"
)

// TemplateNode renders Mustache templates
type TemplateNode struct {
	template string
	field    string
	syntax   string
}

// NewTemplateNode creates a new template node
func NewTemplateNode() *TemplateNode {
	return &TemplateNode{
		field:  "payload",
		syntax: "mustache",
	}
}

// Init initializes the template node with configuration
func (n *TemplateNode) Init(config map[string]interface{}) error {
	if tmpl, ok := config["template"].(string); ok {
		n.template = tmpl
	}
	if field, ok := config["field"].(string); ok && field != "" {
		n.field = field
	}
	if syntax, ok := config["syntax"].(string); ok {
		n.syntax = syntax
	}
	return nil
}

// Execute renders the template with message data
func (n *TemplateNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Build context from message
	context := n.buildContext(msg)

	var result string
	if n.syntax == "plain" {
		result = n.template
	} else {
		result = n.renderMustache(n.template, context)
	}

	// Set result to specified field
	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}

	if n.field == "payload" {
		msg.Payload = map[string]interface{}{"value": result}
	} else {
		msg.Payload[n.field] = result
	}

	return msg, nil
}

// buildContext creates the template context from the message
func (n *TemplateNode) buildContext(msg node.Message) map[string]interface{} {
	context := map[string]interface{}{
		"msg": map[string]interface{}{
			"payload": msg.Payload,
			"topic":   msg.Topic,
		},
	}

	// Add payload fields directly if it's a map
	if msg.Payload != nil {
		context["payload"] = msg.Payload
		for k, v := range msg.Payload {
			context[k] = v
		}
	}

	return context
}

// renderMustache renders a simple Mustache-style template
func (n *TemplateNode) renderMustache(template string, context map[string]interface{}) string {
	result := template

	// Match {{path.to.value}} patterns
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		placeholder := match[0]
		path := strings.TrimSpace(match[1])

		value := n.resolvePath(path, context)
		result = strings.Replace(result, placeholder, fmt.Sprintf("%v", value), 1)
	}

	return result
}

// resolvePath resolves a dot-notation path in the context
func (n *TemplateNode) resolvePath(path string, context map[string]interface{}) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = context

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		switch v := current.(type) {
		case map[string]interface{}:
			if val, ok := v[part]; ok {
				current = val
			} else {
				return ""
			}
		default:
			return ""
		}
	}

	return current
}

// Cleanup cleans up resources
func (n *TemplateNode) Cleanup() error {
	return nil
}
