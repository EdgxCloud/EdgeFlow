package core

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/edgeflow/edgeflow/internal/node"
)

// RegexNode performs regular expression operations
type RegexNode struct {
	pattern     string         // Regex pattern
	compiled    *regexp.Regexp // Compiled pattern
	operation   string         // match, replace, extract, split, test
	replacement string         // Replacement string (for replace)
	property    string         // Property to operate on
	global      bool           // Replace all occurrences
	ignoreCase  bool           // Case insensitive matching
	multiline   bool           // Multiline mode
}

// Init initializes the regex node
func (n *RegexNode) Init(config map[string]interface{}) error {
	// Pattern
	if pattern, ok := config["pattern"].(string); ok {
		n.pattern = pattern
	} else {
		return fmt.Errorf("pattern is required")
	}

	// Operation
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	} else {
		n.operation = "match"
	}

	// Replacement (for replace operation)
	if repl, ok := config["replacement"].(string); ok {
		n.replacement = repl
	}

	// Property to operate on
	if prop, ok := config["property"].(string); ok {
		n.property = prop
	} else {
		n.property = "value"
	}

	// Global flag
	if global, ok := config["global"].(bool); ok {
		n.global = global
	} else {
		n.global = true
	}

	// Ignore case flag
	if ignoreCase, ok := config["ignoreCase"].(bool); ok {
		n.ignoreCase = ignoreCase
	}

	// Multiline flag
	if multiline, ok := config["multiline"].(bool); ok {
		n.multiline = multiline
	}

	// Build pattern with flags
	patternStr := n.pattern
	if n.ignoreCase {
		patternStr = "(?i)" + patternStr
	}
	if n.multiline {
		patternStr = "(?m)" + patternStr
	}

	// Compile pattern
	var err error
	n.compiled, err = regexp.Compile(patternStr)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %w", err)
	}

	return nil
}

// Execute performs the regex operation
func (n *RegexNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get input string
	var input string
	if val, ok := msg.Payload[n.property]; ok {
		input = fmt.Sprintf("%v", val)
	} else if val, ok := msg.Payload["value"]; ok {
		input = fmt.Sprintf("%v", val)
	} else {
		return msg, fmt.Errorf("property %s not found in payload", n.property)
	}

	switch n.operation {
	case "test":
		// Test if pattern matches
		matches := n.compiled.MatchString(input)
		msg.Payload["value"] = matches
		msg.Payload["_regex"] = map[string]interface{}{
			"operation": "test",
			"pattern":   n.pattern,
			"matches":   matches,
		}

	case "match":
		// Find all matches
		var matches []string
		if n.global {
			matches = n.compiled.FindAllString(input, -1)
		} else {
			match := n.compiled.FindString(input)
			if match != "" {
				matches = []string{match}
			}
		}
		msg.Payload["value"] = matches
		msg.Payload["_regex"] = map[string]interface{}{
			"operation": "match",
			"pattern":   n.pattern,
			"count":     len(matches),
		}

	case "extract":
		// Extract groups from matches
		var results []map[string]interface{}
		allMatches := n.compiled.FindAllStringSubmatch(input, -1)
		groupNames := n.compiled.SubexpNames()

		for _, match := range allMatches {
			result := make(map[string]interface{})
			result["full"] = match[0]
			groups := make(map[string]string)
			for i, group := range match[1:] {
				if i+1 < len(groupNames) && groupNames[i+1] != "" {
					groups[groupNames[i+1]] = group
				} else {
					groups[fmt.Sprintf("%d", i+1)] = group
				}
			}
			result["groups"] = groups
			results = append(results, result)

			if !n.global {
				break
			}
		}
		msg.Payload["value"] = results
		msg.Payload["_regex"] = map[string]interface{}{
			"operation": "extract",
			"pattern":   n.pattern,
			"count":     len(results),
		}

	case "replace":
		// Replace matches
		var result string
		if n.global {
			result = n.compiled.ReplaceAllString(input, n.replacement)
		} else {
			// Replace only first occurrence
			match := n.compiled.FindStringIndex(input)
			if match != nil {
				result = input[:match[0]] + n.compiled.ReplaceAllString(input[match[0]:match[1]], n.replacement) + input[match[1]:]
			} else {
				result = input
			}
		}
		msg.Payload["value"] = result
		msg.Payload["_regex"] = map[string]interface{}{
			"operation":   "replace",
			"pattern":     n.pattern,
			"replacement": n.replacement,
			"changed":     result != input,
		}

	case "split":
		// Split by pattern
		parts := n.compiled.Split(input, -1)
		// Remove empty strings if desired
		var filtered []string
		for _, part := range parts {
			if strings.TrimSpace(part) != "" {
				filtered = append(filtered, part)
			}
		}
		msg.Payload["value"] = filtered
		msg.Payload["_regex"] = map[string]interface{}{
			"operation": "split",
			"pattern":   n.pattern,
			"count":     len(filtered),
		}

	default:
		return msg, fmt.Errorf("unknown operation: %s", n.operation)
	}

	return msg, nil
}

// Cleanup releases resources
func (n *RegexNode) Cleanup() error {
	return nil
}

// NewRegexExecutor creates a new regex node executor
func NewRegexExecutor() node.Executor {
	return &RegexNode{}
}

// init registers the regex node
func init() {
	registry := node.GetGlobalRegistry()
	registry.Register(&node.NodeInfo{
		Type:        "regex",
		Name:        "Regex",
		Category:    node.NodeTypeFunction,
		Description: "Regular expression matching, extraction, and replacement",
		Icon:        "search",
		Color:       "#FF8C00",
		Properties: []node.PropertySchema{
			{
				Name:        "pattern",
				Label:       "Pattern",
				Type:        "string",
				Default:     "",
				Required:    true,
				Description: "Regular expression pattern",
			},
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "match",
				Required:    true,
				Description: "Operation to perform",
				Options:     []string{"test", "match", "extract", "replace", "split"},
			},
			{
				Name:        "replacement",
				Label:       "Replacement",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Replacement string (for replace operation, use $1, $2 for groups)",
			},
			{
				Name:        "property",
				Label:       "Property",
				Type:        "string",
				Default:     "value",
				Required:    false,
				Description: "Property containing the input string",
			},
			{
				Name:        "global",
				Label:       "Global",
				Type:        "boolean",
				Default:     true,
				Required:    false,
				Description: "Match/replace all occurrences",
			},
			{
				Name:        "ignoreCase",
				Label:       "Ignore Case",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Case insensitive matching",
			},
			{
				Name:        "multiline",
				Label:       "Multiline",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Multiline mode (^ and $ match line boundaries)",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "string", Description: "String to process"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Result of regex operation"},
		},
		Factory: NewRegexExecutor,
	})
}
