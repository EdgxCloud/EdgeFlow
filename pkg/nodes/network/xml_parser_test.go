package network

import (
	"context"
	"testing"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestXMLParser_Parse tests XML to JSON conversion
func TestXMLParser_Parse(t *testing.T) {
	executor, err := NewXMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		wantKeys []string
		wantErr  bool
	}{
		{
			name: "simple XML",
			input: `<?xml version="1.0"?>
<root>
	<name>Test</name>
	<value>123</value>
</root>`,
			wantKeys: []string{"name", "value"},
			wantErr:  false,
		},
		{
			name: "XML with attributes",
			input: `<person id="1" age="30">
	<name>John</name>
</person>`,
			wantKeys: []string{"@id", "@age", "name"},
			wantErr:  false,
		},
		{
			name: "XML with array",
			input: `<items>
	<item>one</item>
	<item>two</item>
	<item>three</item>
</items>`,
			wantKeys: []string{"item"},
			wantErr:  false,
		},
		{
			name: "XML with CDATA",
			input: `<root>
	<content><![CDATA[This is <b>bold</b> text]]></content>
</root>`,
			wantKeys: []string{"content"},
			wantErr:  false,
		},
		{
			name:     "invalid XML",
			input:    `<root><unclosed>`,
			wantKeys: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{
				Payload: tt.input,
			}

			result, err := executor.Execute(context.Background(), msg)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			payload, ok := result.Payload.(map[string]interface{})
			require.True(t, ok)

			// Check that expected keys are present
			for _, key := range tt.wantKeys {
				assert.Contains(t, payload, key)
			}
		})
	}
}

// TestXMLParser_Stringify tests JSON to XML conversion
func TestXMLParser_Stringify(t *testing.T) {
	executor, err := NewXMLParserExecutor(map[string]interface{}{
		"action": "stringify",
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    interface{}
		contains []string
		wantErr  bool
	}{
		{
			name: "simple object",
			input: map[string]interface{}{
				"name":  "Test",
				"value": 123,
			},
			contains: []string{"<name>Test</name>", "<value>123</value>"},
			wantErr:  false,
		},
		{
			name: "nested object",
			input: map[string]interface{}{
				"person": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			contains: []string{"<person>", "<name>John</name>", "<age>30</age>", "</person>"},
			wantErr:  false,
		},
		{
			name: "array",
			input: map[string]interface{}{
				"items": []interface{}{"one", "two", "three"},
			},
			contains: []string{"<items>", "<item>one</item>", "<item>two</item>", "<item>three</item>", "</items>"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{
				Payload: tt.input,
			}

			result, err := executor.Execute(context.Background(), msg)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			xmlStr, ok := result.Payload.(string)
			require.True(t, ok)

			// Check that expected strings are in the XML
			for _, substr := range tt.contains {
				assert.Contains(t, xmlStr, substr)
			}
		})
	}
}

// TestXMLParser_InvalidConfig tests invalid configuration
func TestXMLParser_InvalidConfig(t *testing.T) {
	_, err := NewXMLParserExecutor(map[string]interface{}{
		"action": "invalid",
	})
	assert.Error(t, err)
}

// TestXMLParser_MissingAction tests missing action
func TestXMLParser_MissingAction(t *testing.T) {
	_, err := NewXMLParserExecutor(map[string]interface{}{})
	assert.Error(t, err)
}

// TestXMLParser_Attributes tests attribute handling
func TestXMLParser_Attributes(t *testing.T) {
	executor, err := NewXMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: `<book id="123" isbn="978-1234567890">
			<title>Test Book</title>
		</book>`,
	}

	result, err := executor.Execute(context.Background(), msg)
	require.NoError(t, err)

	payload, ok := result.Payload.(map[string]interface{})
	require.True(t, ok)

	// Attributes should be prefixed with @
	assert.Contains(t, payload, "@id")
	assert.Contains(t, payload, "@isbn")
	assert.Contains(t, payload, "title")
}

// TestXMLParser_Cleanup tests cleanup
func TestXMLParser_Cleanup(t *testing.T) {
	executor, err := NewXMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}

// BenchmarkXMLParser_Parse benchmarks XML parsing
func BenchmarkXMLParser_Parse(b *testing.B) {
	executor, _ := NewXMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})

	xml := `<root>
		<item>one</item>
		<item>two</item>
		<item>three</item>
	</root>`

	msg := node.Message{Payload: xml}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute(context.Background(), msg)
	}
}

// BenchmarkXMLParser_Stringify benchmarks XML stringification
func BenchmarkXMLParser_Stringify(b *testing.B) {
	executor, _ := NewXMLParserExecutor(map[string]interface{}{
		"action": "stringify",
	})

	data := map[string]interface{}{
		"items": []interface{}{"one", "two", "three"},
	}

	msg := node.Message{Payload: data}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute(context.Background(), msg)
	}
}
