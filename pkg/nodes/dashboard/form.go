package dashboard

import (
	"context"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// FormField represents a form field
type FormField struct {
	Name        string      `json:"name"`
	Label       string      `json:"label"`
	Type        string      `json:"type"` // text, number, email, password, checkbox, radio
	Required    bool        `json:"required"`
	Placeholder string      `json:"placeholder,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Options     []string    `json:"options,omitempty"` // For radio/select
}

// FormNode provides form builder functionality
type FormNode struct {
	*BaseWidget
	fields      []FormField
	submitLabel string
	resetLabel  string
	formData    map[string]interface{}
}

// NewFormNode creates a new form widget
func NewFormNode() *FormNode {
	return &FormNode{
		BaseWidget:  NewBaseWidget(WidgetTypeForm),
		fields:      make([]FormField, 0),
		submitLabel: "Submit",
		resetLabel:  "Reset",
		formData:    make(map[string]interface{}),
	}
}

// Init initializes the form node
func (n *FormNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	if submitLabel, ok := config["submitLabel"].(string); ok {
		n.submitLabel = submitLabel
	}
	if resetLabel, ok := config["resetLabel"].(string); ok {
		n.resetLabel = resetLabel
	}

	// Parse form fields
	if fields, ok := config["fields"].([]interface{}); ok {
		for _, f := range fields {
			if fieldMap, ok := f.(map[string]interface{}); ok {
				field := FormField{
					Name:        fieldMap["name"].(string),
					Label:       fieldMap["label"].(string),
					Type:        fieldMap["type"].(string),
					Required:    false,
					Placeholder: "",
				}
				if req, ok := fieldMap["required"].(bool); ok {
					field.Required = req
				}
				if ph, ok := fieldMap["placeholder"].(string); ok {
					field.Placeholder = ph
				}
				if def, ok := fieldMap["default"]; ok {
					field.Default = def
				}
				if opts, ok := fieldMap["options"].([]interface{}); ok {
					field.Options = make([]string, len(opts))
					for i, opt := range opts {
						field.Options[i] = opt.(string)
					}
				}
				n.fields = append(n.fields, field)
			}
		}
	}

	return nil
}

// Execute handles form submission
func (n *FormNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract form data from user interaction
	if formData, ok := msg.Payload["formData"].(map[string]interface{}); ok {
		n.formData = formData

		// Send output message with form data
		outputMsg := node.Message{
			Type:    node.MessageTypeData,
			Payload: formData,
		}
		n.SendOutput(outputMsg)
	}

	return node.Message{}, nil
}

// SetManager sets the dashboard manager
func (n *FormNode) SetManager(manager *Manager) {
	n.manager = manager
}
