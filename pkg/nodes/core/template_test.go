package core

import (
	"context"
	"testing"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateNode_Init(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid template",
			config: map[string]interface{}{
				"template": "Hello {{msg.payload.name}}",
				"syntax":   "mustache",
				"field":    "payload",
			},
			wantErr: false,
		},
		{
			name: "empty template",
			config: map[string]interface{}{
				"template": "",
			},
			wantErr: true,
		},
		{
			name: "plain text mode",
			config: map[string]interface{}{
				"template": "Static text",
				"syntax":   "plain",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewTemplateNode()
			err := n.Init(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTemplateNode_Execute(t *testing.T) {
	tests := []struct {
		name         string
		config       map[string]interface{}
		inputMsg     node.Message
		expectedOut  interface{}
		expectedErr  bool
	}{
		{
			name: "simple mustache template",
			config: map[string]interface{}{
				"template": "Hello {{msg.payload.name}}!",
				"syntax":   "mustache",
				"field":    "payload",
			},
			inputMsg: node.Message{
				Payload: map[string]interface{}{
					"name": "World",
				},
			},
			expectedOut: "Hello World!",
			expectedErr: false,
		},
		{
			name: "topic output",
			config: map[string]interface{}{
				"template": "sensor/{{msg.payload.id}}/temperature",
				"syntax":   "mustache",
				"field":    "topic",
			},
			inputMsg: node.Message{
				Payload: map[string]interface{}{
					"id": "123",
				},
			},
			expectedOut: "sensor/123/temperature",
			expectedErr: false,
		},
		{
			name: "plain text mode",
			config: map[string]interface{}{
				"template": "This is static text",
				"syntax":   "plain",
				"field":    "payload",
			},
			inputMsg: node.Message{
				Payload: map[string]interface{}{
					"ignored": "value",
				},
			},
			expectedOut: "This is static text",
			expectedErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewTemplateNode()
			err := n.Init(tt.config)
			require.NoError(t, err)

			resultMsg, err := n.Execute(context.Background(), tt.inputMsg)
			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check output based on field
				field := tt.config["field"].(string)
				switch field {
				case "payload":
					assert.Equal(t, tt.expectedOut, resultMsg.Payload)
				case "topic":
					assert.Equal(t, tt.expectedOut, resultMsg.Topic)
				}
			}
		})
	}
}

func TestTemplateNode_ConvertMustacheToGo(t *testing.T) {
	n := NewTemplateNode()

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "{{msg.payload}}",
			expected: "{{.Msg.Payload}}",
		},
		{
			input:    "Hello {{msg.payload.name}}!",
			expected: "Hello {{.Msg.Payload.Name}}!",
		},
		{
			input:    "{{msg.topic}} - {{msg.payload.value}}",
			expected: "{{.Msg.Topic}} - {{.Msg.Payload.Value}}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := n.convertMustacheToGo(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTemplateNode_Cleanup(t *testing.T) {
	n := NewTemplateNode()
	err := n.Cleanup()
	assert.NoError(t, err)
}
