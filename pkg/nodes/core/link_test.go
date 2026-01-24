package core

import (
	"context"
	"testing"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinkInOut_BasicFlow(t *testing.T) {
	// Clear registry before test
	ClearLinkRegistry()

	// Create Link In node
	linkIn := NewLinkInNode()
	err := linkIn.Init(map[string]interface{}{
		"linkId": "test-link-1",
		"scope":  "global",
	})
	require.NoError(t, err)

	// Start Link In node
	ctx := context.Background()
	err = linkIn.Start(ctx)
	require.NoError(t, err)

	// Create Link Out node
	linkOut := NewLinkOutNode()
	err = linkOut.Init(map[string]interface{}{
		"linkId": "test-link-1",
		"scope":  "global",
		"mode":   "continue",
	})
	require.NoError(t, err)

	// Send message through Link Out
	testMsg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"test": "value",
			"num":  42,
		},
		Topic: "test/topic",
	}

	resultMsg, err := linkOut.Execute(ctx, testMsg)
	require.NoError(t, err)

	// Verify message continues through (mode=continue)
	assert.Equal(t, testMsg.Payload, resultMsg.Payload)

	// Check that Link In received the message
	outputChan := linkIn.GetOutputChannel()
	select {
	case receivedMsg := <-outputChan:
		assert.Equal(t, testMsg.Payload, receivedMsg.Payload)
		assert.Equal(t, testMsg.Topic, receivedMsg.Topic)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive message from Link In")
	}

	// Cleanup
	linkIn.Cleanup()
	linkOut.Cleanup()
}

func TestLinkInOut_ReturnMode(t *testing.T) {
	// Clear registry
	ClearLinkRegistry()

	// Create Link In and Link Out with return mode
	linkIn := NewLinkInNode()
	err := linkIn.Init(map[string]interface{}{
		"linkId": "test-link-return",
		"scope":  "global",
	})
	require.NoError(t, err)
	err = linkIn.Start(context.Background())
	require.NoError(t, err)

	linkOut := NewLinkOutNode()
	err = linkOut.Init(map[string]interface{}{
		"linkId": "test-link-return",
		"scope":  "global",
		"mode":   "return", // Return mode - don't continue to output
	})
	require.NoError(t, err)

	// Send message
	testMsg := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"data": "test"},
	}

	resultMsg, err := linkOut.Execute(context.Background(), testMsg)
	require.NoError(t, err)

	// Verify message does NOT continue through (return mode)
	assert.Empty(t, resultMsg.Payload)

	// Link In should still receive it
	select {
	case receivedMsg := <-linkIn.GetOutputChannel():
		assert.Equal(t, testMsg.Payload, receivedMsg.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Link In did not receive message")
	}

	// Cleanup
	linkIn.Cleanup()
	linkOut.Cleanup()
}

func TestLinkInOut_MultipleTargets(t *testing.T) {
	// Clear registry
	ClearLinkRegistry()

	// Create multiple Link In nodes
	linkIn1 := NewLinkInNode()
	err := linkIn1.Init(map[string]interface{}{
		"linkId": "multi-link",
		"scope":  "global",
	})
	require.NoError(t, err)
	err = linkIn1.Start(context.Background())
	require.NoError(t, err)

	linkIn2 := NewLinkInNode()
	err = linkIn2.Init(map[string]interface{}{
		"linkId": "multi-link",
		"scope":  "global",
	})
	require.NoError(t, err)
	err = linkIn2.Start(context.Background())
	require.NoError(t, err)

	// Create Link Out targeting both
	linkOut := NewLinkOutNode()
	err = linkOut.Init(map[string]interface{}{
		"linkId": "multi-link",
		"scope":  "global",
	})
	require.NoError(t, err)

	// Send message
	testMsg := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"broadcast": true},
	}

	_, err = linkOut.Execute(context.Background(), testMsg)
	require.NoError(t, err)

	// Both Link In nodes should receive the message
	select {
	case msg1 := <-linkIn1.GetOutputChannel():
		assert.Equal(t, testMsg.Payload, msg1.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Link In 1 did not receive message")
	}

	select {
	case msg2 := <-linkIn2.GetOutputChannel():
		assert.Equal(t, testMsg.Payload, msg2.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Link In 2 did not receive message")
	}

	// Cleanup
	linkIn1.Cleanup()
	linkIn2.Cleanup()
	linkOut.Cleanup()
}

func TestLinkInOut_FlowScope(t *testing.T) {
	// Clear registry
	ClearLinkRegistry()

	// Create flow-scoped Link In
	linkIn := NewLinkInNode()
	err := linkIn.Init(map[string]interface{}{
		"linkId": "flow-link",
		"scope":  "flow",
		"flowId": "flow-123",
	})
	require.NoError(t, err)
	err = linkIn.Start(context.Background())
	require.NoError(t, err)

	// Create Link Out in same flow
	linkOut1 := NewLinkOutNode()
	err = linkOut1.Init(map[string]interface{}{
		"linkId": "flow-link",
		"scope":  "flow",
		"flowId": "flow-123",
	})
	require.NoError(t, err)

	// Create Link Out in different flow
	linkOut2 := NewLinkOutNode()
	err = linkOut2.Init(map[string]interface{}{
		"linkId": "flow-link",
		"scope":  "flow",
		"flowId": "flow-456", // Different flow
	})
	require.NoError(t, err)

	// Send from same flow - should work
	testMsg1 := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"same": "flow"},
	}
	_, err = linkOut1.Execute(context.Background(), testMsg1)
	require.NoError(t, err)

	select {
	case receivedMsg := <-linkIn.GetOutputChannel():
		assert.Equal(t, testMsg1.Payload, receivedMsg.Payload)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Link In did not receive message from same flow")
	}

	// Send from different flow - should NOT work
	testMsg2 := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"different": "flow"},
	}
	_, err = linkOut2.Execute(context.Background(), testMsg2)
	require.NoError(t, err)

	// Should NOT receive message from different flow
	select {
	case <-linkIn.GetOutputChannel():
		t.Fatal("Link In should not receive message from different flow")
	case <-time.After(100 * time.Millisecond):
		// Expected - no message
	}

	// Cleanup
	linkIn.Cleanup()
	linkOut1.Cleanup()
	linkOut2.Cleanup()
}

func TestLinkRegistry_Stats(t *testing.T) {
	// Clear registry
	ClearLinkRegistry()

	// Create multiple Link In nodes
	linkIn1 := NewLinkInNode()
	linkIn1.Init(map[string]interface{}{"linkId": "link-a", "scope": "global"})
	linkIn1.Start(context.Background())

	linkIn2 := NewLinkInNode()
	linkIn2.Init(map[string]interface{}{"linkId": "link-a", "scope": "global"})
	linkIn2.Start(context.Background())

	linkIn3 := NewLinkInNode()
	linkIn3.Init(map[string]interface{}{"linkId": "link-b", "scope": "global"})
	linkIn3.Start(context.Background())

	// Check stats
	stats := GetLinkStats()
	assert.Equal(t, 2, stats["link-a"])
	assert.Equal(t, 1, stats["link-b"])

	// Cleanup one node
	linkIn1.Cleanup()

	// Check stats again
	stats = GetLinkStats()
	assert.Equal(t, 1, stats["link-a"])
	assert.Equal(t, 1, stats["link-b"])

	// Cleanup remaining
	linkIn2.Cleanup()
	linkIn3.Cleanup()

	// Stats should be empty
	stats = GetLinkStats()
	assert.Equal(t, 0, len(stats))
}
