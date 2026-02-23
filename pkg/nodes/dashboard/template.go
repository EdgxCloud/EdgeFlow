package dashboard

import (
	"bytes"
	"context"
	"html/template"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// TemplateNode displays custom HTML/template content
type TemplateNode struct {
	*BaseWidget
	templateStr string
	template    *template.Template
	allowHTML   bool
	data        map[string]interface{}
}

// NewTemplateNode creates a new template widget
func NewTemplateNode() *TemplateNode {
	return &TemplateNode{
		BaseWidget:  NewBaseWidget(WidgetTypeTemplate),
		templateStr: "<p>{{.value}}</p>",
		allowHTML:   false,
		data:        make(map[string]interface{}),
	}
}

// Init initializes the template node
func (n *TemplateNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	if templateStr, ok := config["template"].(string); ok {
		n.templateStr = templateStr

		// Parse template
		tmpl, err := template.New("widget").Parse(templateStr)
		if err != nil {
			return err
		}
		n.template = tmpl
	}

	if allowHTML, ok := config["allowHTML"].(bool); ok {
		n.allowHTML = allowHTML
	}

	return nil
}

// Execute processes incoming messages and renders template
func (n *TemplateNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Update data with message payload
	for k, v := range msg.Payload {
		n.data[k] = v
	}

	// Render template
	var buf bytes.Buffer
	if n.template != nil {
		if err := n.template.Execute(&buf, n.data); err != nil {
			return node.Message{}, err
		}

		// Update dashboard with rendered content
		if n.manager != nil {
			templateData := map[string]interface{}{
				"content":   buf.String(),
				"allowHTML": n.allowHTML,
			}
			n.manager.UpdateWidget(n.id, templateData)
		}
	}

	return node.Message{}, nil
}

// SetManager sets the dashboard manager
func (n *TemplateNode) SetManager(manager *Manager) {
	n.manager = manager
}
