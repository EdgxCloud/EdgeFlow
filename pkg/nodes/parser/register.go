package parser

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterNodes registers all parser nodes with the registry
func RegisterNodes(registry *node.Registry) error {
	// Register HTML Parser Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "html",
		Name:        "HTML Parser",
		Category:    node.NodeTypeFunction,
		Description: "Parse HTML and extract elements with CSS selectors",
		Icon:        "code",
		Color:       "#e34c26",
		Properties: []node.PropertySchema{
			{
				Name:        "selector",
				Label:       "CSS Selector",
				Type:        "string",
				Default:     "",
				Required:    true,
				Description: "CSS selector (tag, .class, #id)",
			},
			{
				Name:        "output",
				Label:       "Output",
				Type:        "select",
				Default:     "text",
				Options:     []string{"text", "html", "attr"},
				Description: "Output type",
			},
			{
				Name:        "attr",
				Label:       "Attribute",
				Type:        "string",
				Default:     "",
				Description: "Attribute name when output is attr",
			},
			{
				Name:        "multiple",
				Label:       "Multiple Results",
				Type:        "boolean",
				Default:     false,
				Description: "Return all matches as array",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "string", Description: "HTML content"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Extracted content"},
		},
		Factory: func() node.Executor {
			return NewHTMLNode()
		},
	}); err != nil {
		return err
	}

	return nil
}
