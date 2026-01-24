package parser

import (
	"context"
	"fmt"
	"strings"

	"github.com/edgeflow/edgeflow/internal/node"
	"golang.org/x/net/html"
)

// HTMLNode parses HTML and extracts elements using CSS-like selectors
type HTMLNode struct {
	selector string // CSS selector (simplified: tag, .class, #id)
	output   string // "text", "html", "attr"
	attr     string // Attribute name when output is "attr"
	multiple bool   // Return all matches or just first
}

// NewHTMLNode creates a new HTML parser node
func NewHTMLNode() *HTMLNode {
	return &HTMLNode{
		selector: "",
		output:   "text",
		attr:     "",
		multiple: false,
	}
}

// Init initializes the HTML node with configuration
func (n *HTMLNode) Init(config map[string]interface{}) error {
	if selector, ok := config["selector"].(string); ok {
		n.selector = selector
	}
	if output, ok := config["output"].(string); ok {
		n.output = output
	}
	if attr, ok := config["attr"].(string); ok {
		n.attr = attr
	}
	if multiple, ok := config["multiple"].(bool); ok {
		n.multiple = multiple
	}
	return nil
}

// Execute parses HTML and extracts elements
func (n *HTMLNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if msg.Payload == nil {
		return msg, fmt.Errorf("html node: nil payload")
	}

	// Get HTML string from payload
	var htmlStr string
	if val, ok := msg.Payload["value"].(string); ok {
		htmlStr = val
	} else if val, ok := msg.Payload["html"].(string); ok {
		htmlStr = val
	} else {
		// Try to convert payload to string
		for _, v := range msg.Payload {
			if s, ok := v.(string); ok {
				htmlStr = s
				break
			}
		}
	}

	if htmlStr == "" {
		return msg, fmt.Errorf("html node: no HTML content found")
	}

	// Parse HTML
	doc, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		return msg, fmt.Errorf("html node: parse error: %w", err)
	}

	// Find matching elements
	matches := n.findElements(doc)

	// Extract output based on mode
	var result interface{}
	if n.multiple {
		results := make([]string, 0, len(matches))
		for _, match := range matches {
			results = append(results, n.extractOutput(match))
		}
		result = results
	} else if len(matches) > 0 {
		result = n.extractOutput(matches[0])
	} else {
		result = ""
	}

	msg.Payload = map[string]interface{}{
		"value":    result,
		"selector": n.selector,
		"count":    len(matches),
	}

	return msg, nil
}

// findElements finds elements matching the selector
func (n *HTMLNode) findElements(doc *html.Node) []*html.Node {
	var matches []*html.Node

	// Parse selector (simplified CSS selector support)
	selectorType, selectorValue := n.parseSelector()

	var traverse func(*html.Node)
	traverse = func(node *html.Node) {
		if node.Type == html.ElementNode {
			if n.matchesSelector(node, selectorType, selectorValue) {
				matches = append(matches, node)
				if !n.multiple {
					return // Stop after first match if not multiple
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
			if !n.multiple && len(matches) > 0 {
				return
			}
		}
	}

	traverse(doc)
	return matches
}

// parseSelector parses the CSS selector into type and value
func (n *HTMLNode) parseSelector() (string, string) {
	selector := strings.TrimSpace(n.selector)
	if selector == "" {
		return "", ""
	}

	if strings.HasPrefix(selector, "#") {
		return "id", strings.TrimPrefix(selector, "#")
	}
	if strings.HasPrefix(selector, ".") {
		return "class", strings.TrimPrefix(selector, ".")
	}
	if strings.Contains(selector, "[") {
		// Attribute selector: tag[attr=value]
		parts := strings.SplitN(selector, "[", 2)
		return "tag", parts[0]
	}
	return "tag", selector
}

// matchesSelector checks if a node matches the selector
func (n *HTMLNode) matchesSelector(node *html.Node, selectorType, selectorValue string) bool {
	switch selectorType {
	case "tag":
		return node.Data == selectorValue
	case "id":
		for _, attr := range node.Attr {
			if attr.Key == "id" && attr.Val == selectorValue {
				return true
			}
		}
	case "class":
		for _, attr := range node.Attr {
			if attr.Key == "class" {
				classes := strings.Fields(attr.Val)
				for _, class := range classes {
					if class == selectorValue {
						return true
					}
				}
			}
		}
	}
	return false
}

// extractOutput extracts the desired output from a node
func (n *HTMLNode) extractOutput(node *html.Node) string {
	switch n.output {
	case "text":
		return n.extractText(node)
	case "html":
		return n.extractHTML(node)
	case "attr":
		return n.extractAttr(node, n.attr)
	default:
		return n.extractText(node)
	}
}

// extractText extracts text content from a node
func (n *HTMLNode) extractText(node *html.Node) string {
	var text strings.Builder

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			text.WriteString(n.Data)
		}
		for child := n.FirstChild; child != nil; child = child.NextSibling {
			traverse(child)
		}
	}

	traverse(node)
	return strings.TrimSpace(text.String())
}

// extractHTML extracts outer HTML from a node
func (n *HTMLNode) extractHTML(node *html.Node) string {
	var buf strings.Builder
	html.Render(&buf, node)
	return buf.String()
}

// extractAttr extracts an attribute value from a node
func (n *HTMLNode) extractAttr(node *html.Node, attrName string) string {
	for _, attr := range node.Attr {
		if attr.Key == attrName {
			return attr.Val
		}
	}
	return ""
}

// Cleanup cleans up resources
func (n *HTMLNode) Cleanup() error {
	return nil
}
