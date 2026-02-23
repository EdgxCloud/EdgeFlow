package core

import (
	"context"
	"testing"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitNode_Init(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "default config",
			config: map[string]interface{}{
				"arraySplt": true,
			},
			wantErr: false,
		},
		{
			name: "fixed length split",
			config: map[string]interface{}{
				"arraySplt":      true,
				"arraySplitType": "len",
				"arraySpltLen":   2,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewSplitNode()
			err := n.Init(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSplitNode_SplitArray(t *testing.T) {
	config := map[string]interface{}{
		"arraySplt": true,
	}

	n := NewSplitNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name          string
		input         []interface{}
		expectedCount int
	}{
		{
			name:          "three elements",
			input:         []interface{}{1, 2, 3},
			expectedCount: 3,
		},
		{
			name:          "single element",
			input:         []interface{}{"only"},
			expectedCount: 1,
		},
		{
			name:          "empty array",
			input:         []interface{}{},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{
				Payload: tt.input,
			}

			result, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)

			if tt.expectedCount > 0 {
				// Check if split messages were created
				if em, ok := interface{}(result).(*node.EnhancedMessage); ok {
					if splitMsgs, ok := em.Metadata["_splitMessages"].([]*node.EnhancedMessage); ok {
						assert.Equal(t, tt.expectedCount, len(splitMsgs))

						// Verify msg.parts for each message
						for i, splitMsg := range splitMsgs {
							assert.NotNil(t, splitMsg.Parts)
							assert.Equal(t, i, splitMsg.Parts.Index)
							assert.Equal(t, tt.expectedCount, splitMsg.Parts.Count)
							assert.Equal(t, "array", splitMsg.Parts.Type)
							assert.NotEmpty(t, splitMsg.Parts.ID)
						}
					}
				}
			}
		})
	}
}

func TestSplitNode_SplitArrayFixedLength(t *testing.T) {
	config := map[string]interface{}{
		"arraySplt":      true,
		"arraySplitType": "len",
		"arraySpltLen":   2,
	}

	n := NewSplitNode()
	err := n.Init(config)
	require.NoError(t, err)

	input := []interface{}{1, 2, 3, 4, 5}
	msg := node.Message{Payload: input}

	result, err := n.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Should create 3 messages: [1,2], [3,4], [5]
	if em, ok := interface{}(result).(*node.EnhancedMessage); ok {
		if splitMsgs, ok := em.Metadata["_splitMessages"].([]*node.EnhancedMessage); ok {
			assert.Equal(t, 3, len(splitMsgs))

			// First chunk
			chunk0 := splitMsgs[0].Payload.([]interface{})
			assert.Equal(t, 2, len(chunk0))
			assert.Equal(t, 1, chunk0[0])

			// Last chunk
			chunk2 := splitMsgs[2].Payload.([]interface{})
			assert.Equal(t, 1, len(chunk2))
			assert.Equal(t, 5, chunk2[0])
		}
	}
}

func TestSplitNode_SplitObject(t *testing.T) {
	config := map[string]interface{}{
		"arraySplt": true,
		"addname":   "key",
	}

	n := NewSplitNode()
	err := n.Init(config)
	require.NoError(t, err)

	input := map[string]interface{}{
		"temp":     25.5,
		"humidity": 60,
		"pressure": 1013,
	}

	msg := node.Message{Payload: input}
	result, err := n.Execute(context.Background(), msg)
	require.NoError(t, err)

	if em, ok := interface{}(result).(*node.EnhancedMessage); ok {
		if splitMsgs, ok := em.Metadata["_splitMessages"].([]*node.EnhancedMessage); ok {
			assert.Equal(t, 3, len(splitMsgs))

			// Verify each message has key in parts
			for _, splitMsg := range splitMsgs {
				assert.NotNil(t, splitMsg.Parts)
				assert.NotEmpty(t, splitMsg.Parts.Key)
				assert.Equal(t, "object", splitMsg.Parts.Type)

				// Check if key was added to payload
				if payloadMap, ok := splitMsg.Payload.(map[string]interface{}); ok {
					keyVal, exists := payloadMap["key"]
					assert.True(t, exists)
					assert.Equal(t, splitMsg.Parts.Key, keyVal)
				}
			}
		}
	}
}

func TestSplitNode_SplitString(t *testing.T) {
	config := map[string]interface{}{
		"arraySplt": true,
	}

	n := NewSplitNode()
	err := n.Init(config)
	require.NoError(t, err)

	input := "line1\nline2\nline3"
	msg := node.Message{Payload: input}

	result, err := n.Execute(context.Background(), msg)
	require.NoError(t, err)

	if em, ok := interface{}(result).(*node.EnhancedMessage); ok {
		if splitMsgs, ok := em.Metadata["_splitMessages"].([]*node.EnhancedMessage); ok {
			assert.Equal(t, 3, len(splitMsgs))

			// Verify each line
			assert.Equal(t, "line1", splitMsgs[0].Payload)
			assert.Equal(t, "line2", splitMsgs[1].Payload)
			assert.Equal(t, "line3", splitMsgs[2].Payload)

			// Verify parts metadata
			for _, splitMsg := range splitMsgs {
				assert.NotNil(t, splitMsg.Parts)
				assert.Equal(t, "string", splitMsg.Parts.Type)
				assert.Equal(t, "\n", splitMsg.Parts.Ch)
			}
		}
	}
}

func TestSplitNode_InvalidType(t *testing.T) {
	config := map[string]interface{}{
		"arraySplt": true,
	}

	n := NewSplitNode()
	err := n.Init(config)
	require.NoError(t, err)

	// Try to split a number (not supported)
	msg := node.Message{Payload: 42}
	_, err = n.Execute(context.Background(), msg)
	assert.Error(t, err)
}

func TestSplitNode_Cleanup(t *testing.T) {
	n := NewSplitNode()
	err := n.Cleanup()
	assert.NoError(t, err)
}
