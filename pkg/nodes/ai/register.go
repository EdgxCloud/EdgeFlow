package ai

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterAllNodes ثبت تمام نودهای AI
func RegisterAllNodes(registry *node.Registry) error {
	// Register OpenAI Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "openai",
		Name:        "OpenAI",
		Category:    node.NodeTypeFunction,
		Description: "تولید متن با OpenAI GPT",
		Icon:        "cpu",
		Color:       "#10a37f",
		Properties: []node.PropertySchema{
			{
				Name:        "apiKey",
				Label:       "API Key",
				Type:        "string",
				Default:     "",
				Required:    true,
				Description: "OpenAI API key",
			},
			{
				Name:        "model",
				Label:       "Model",
				Type:        "select",
				Default:     "gpt-3.5-turbo",
				Required:    false,
				Options:     []string{"gpt-4", "gpt-4-turbo", "gpt-3.5-turbo"},
				Description: "OpenAI model",
			},
			{
				Name:        "temperature",
				Label:       "Temperature",
				Type:        "number",
				Default:     0.7,
				Required:    false,
				Description: "Temperature (0-2)",
			},
			{
				Name:        "maxTokens",
				Label:       "Max Tokens",
				Type:        "number",
				Default:     1000,
				Required:    false,
				Description: "Maximum tokens",
			},
		},
		Inputs: []node.PortSchema{
			{
				Name:        "input",
				Label:       "Input",
				Type:        "any",
				Description: "Message input (prompt, system, etc.)",
			},
		},
		Outputs: []node.PortSchema{
			{
				Name:        "output",
				Label:       "Output",
				Type:        "any",
				Description: "AI response",
			},
		},
		Factory: func() node.Executor {
			return NewOpenAIExecutor()
		},
	}); err != nil {
		return err
	}

	// Register Anthropic Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "anthropic",
		Name:        "Anthropic Claude",
		Category:    node.NodeTypeFunction,
		Description: "تولید متن با Anthropic Claude",
		Icon:        "cpu",
		Color:       "#cc785c",
		Properties: []node.PropertySchema{
			{
				Name:        "apiKey",
				Label:       "API Key",
				Type:        "string",
				Default:     "",
				Required:    true,
				Description: "Anthropic API key",
			},
			{
				Name:        "model",
				Label:       "Model",
				Type:        "select",
				Default:     "claude-3-5-sonnet-20241022",
				Required:    false,
				Options:     []string{"claude-3-5-sonnet-20241022", "claude-3-opus-20240229", "claude-3-sonnet-20240229"},
				Description: "Claude model",
			},
			{
				Name:        "temperature",
				Label:       "Temperature",
				Type:        "number",
				Default:     1.0,
				Required:    false,
				Description: "Temperature (0-1)",
			},
			{
				Name:        "maxTokens",
				Label:       "Max Tokens",
				Type:        "number",
				Default:     1024,
				Required:    false,
				Description: "Maximum tokens",
			},
		},
		Inputs: []node.PortSchema{
			{
				Name:        "input",
				Label:       "Input",
				Type:        "any",
				Description: "Message input (prompt, system, etc.)",
			},
		},
		Outputs: []node.PortSchema{
			{
				Name:        "output",
				Label:       "Output",
				Type:        "any",
				Description: "AI response",
			},
		},
		Factory: func() node.Executor {
			return NewAnthropicExecutor()
		},
	}); err != nil {
		return err
	}

	// Register Ollama Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "ollama",
		Name:        "Ollama",
		Category:    node.NodeTypeFunction,
		Description: "تولید متن با Ollama (Local LLM)",
		Icon:        "cpu",
		Color:       "#000000",
		Properties: []node.PropertySchema{
			{
				Name:        "baseUrl",
				Label:       "Base URL",
				Type:        "string",
				Default:     "http://localhost:11434",
				Required:    false,
				Description: "Ollama server URL",
			},
			{
				Name:        "model",
				Label:       "Model",
				Type:        "select",
				Default:     "gemma3:1b",
				Required:    false,
				Options:     []string{"gemma3:1b", "gemma3:270m", "gemma3:4b", "tinyllama", "phi3:mini", "qwen2:0.5b", "llama3:8b"},
				Description: "Ollama model (gemma3:1b recommended for Raspberry Pi)",
			},
			{
				Name:        "temperature",
				Label:       "Temperature",
				Type:        "number",
				Default:     0.7,
				Required:    false,
				Description: "Temperature (0-2)",
			},
			{
				Name:        "stream",
				Label:       "Stream",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Stream response",
			},
		},
		Inputs: []node.PortSchema{
			{
				Name:        "input",
				Label:       "Input",
				Type:        "any",
				Description: "Message input (prompt, system, etc.)",
			},
		},
		Outputs: []node.PortSchema{
			{
				Name:        "output",
				Label:       "Output",
				Type:        "any",
				Description: "AI response",
			},
		},
		Factory: func() node.Executor {
			return NewOllamaExecutor()
		},
	}); err != nil {
		return err
	}

	return nil
}
