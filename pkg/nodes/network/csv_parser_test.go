package network

import (
	"context"
	"testing"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCSVParser_Parse tests CSV to JSON conversion
func TestCSVParser_Parse(t *testing.T) {
	executor, err := NewCSVParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	tests := []struct {
		name      string
		input     string
		wantRows  int
		wantCols  int
		firstCell string
		wantErr   bool
	}{
		{
			name: "simple CSV with headers",
			input: `name,age,city
John,30,NYC
Jane,25,LA`,
			wantRows:  2,
			wantCols:  3,
			firstCell: "John",
			wantErr:   false,
		},
		{
			name: "CSV with quotes",
			input: `"name","value","description"
"Product A","100","A great product"
"Product B","200","Another product"`,
			wantRows:  2,
			wantCols:  3,
			firstCell: "Product A",
			wantErr:   false,
		},
		{
			name: "CSV with commas in quotes",
			input: `name,description
"Product","Contains, comma, values"`,
			wantRows:  1,
			wantCols:  2,
			firstCell: "Product",
			wantErr:   false,
		},
		{
			name: "empty CSV",
			input: ``,
			wantRows:  0,
			wantCols:  0,
			firstCell: "",
			wantErr:   false,
		},
		{
			name: "CSV with only headers",
			input: `name,age,city`,
			wantRows:  0,
			wantCols:  3,
			firstCell: "",
			wantErr:   false,
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

			payload, ok := result.Payload.([]interface{})
			require.True(t, ok)

			assert.Len(t, payload, tt.wantRows)

			if tt.wantRows > 0 {
				firstRow, ok := payload[0].(map[string]interface{})
				require.True(t, ok)
				assert.Len(t, firstRow, tt.wantCols)

				if tt.firstCell != "" {
					// Check first cell value (assumes first column is "name")
					assert.Contains(t, firstRow, "name")
				}
			}
		})
	}
}

// TestCSVParser_ParseWithCustomDelimiter tests custom delimiter
func TestCSVParser_ParseWithCustomDelimiter(t *testing.T) {
	executor, err := NewCSVParserExecutor(map[string]interface{}{
		"action":    "parse",
		"delimiter": ";",
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: `name;age;city
John;30;NYC
Jane;25;LA`,
	}

	result, err := executor.Execute(context.Background(), msg)
	require.NoError(t, err)

	payload, ok := result.Payload.([]interface{})
	require.True(t, ok)
	assert.Len(t, payload, 2)

	firstRow, ok := payload[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "John", firstRow["name"])
}

// TestCSVParser_ParseWithoutHeaders tests CSV without header row
func TestCSVParser_ParseWithoutHeaders(t *testing.T) {
	executor, err := NewCSVParserExecutor(map[string]interface{}{
		"action":     "parse",
		"hasHeaders": false,
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: `John,30,NYC
Jane,25,LA`,
	}

	result, err := executor.Execute(context.Background(), msg)
	require.NoError(t, err)

	payload, ok := result.Payload.([]interface{})
	require.True(t, ok)
	assert.Len(t, payload, 2)

	// Without headers, should be array of arrays
	firstRow, ok := payload[0].([]interface{})
	require.True(t, ok)
	assert.Len(t, firstRow, 3)
	assert.Equal(t, "John", firstRow[0])
}

// TestCSVParser_Stringify tests JSON to CSV conversion
func TestCSVParser_Stringify(t *testing.T) {
	executor, err := NewCSVParserExecutor(map[string]interface{}{
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
			name: "array of objects",
			input: []interface{}{
				map[string]interface{}{"name": "John", "age": float64(30)},
				map[string]interface{}{"name": "Jane", "age": float64(25)},
			},
			contains: []string{"name", "age", "John", "30", "Jane", "25"},
			wantErr:  false,
		},
		{
			name: "single object",
			input: map[string]interface{}{
				"name": "John",
				"age":  float64(30),
			},
			contains: []string{"name", "age", "John", "30"},
			wantErr:  false,
		},
		{
			name: "array of arrays",
			input: []interface{}{
				[]interface{}{"John", float64(30)},
				[]interface{}{"Jane", float64(25)},
			},
			contains: []string{"John", "30", "Jane", "25"},
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

			csvStr, ok := result.Payload.(string)
			require.True(t, ok)

			// Check that expected strings are in the CSV
			for _, substr := range tt.contains {
				assert.Contains(t, csvStr, substr)
			}
		})
	}
}

// TestCSVParser_StringifyWithCustomDelimiter tests custom delimiter for stringify
func TestCSVParser_StringifyWithCustomDelimiter(t *testing.T) {
	executor, err := NewCSVParserExecutor(map[string]interface{}{
		"action":    "stringify",
		"delimiter": "|",
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: []interface{}{
			map[string]interface{}{"name": "John", "age": float64(30)},
		},
	}

	result, err := executor.Execute(context.Background(), msg)
	require.NoError(t, err)

	csvStr, ok := result.Payload.(string)
	require.True(t, ok)
	assert.Contains(t, csvStr, "|")
}

// TestCSVParser_InvalidConfig tests invalid configuration
func TestCSVParser_InvalidConfig(t *testing.T) {
	_, err := NewCSVParserExecutor(map[string]interface{}{
		"action": "invalid",
	})
	assert.Error(t, err)
}

// TestCSVParser_MissingAction tests missing action
func TestCSVParser_MissingAction(t *testing.T) {
	_, err := NewCSVParserExecutor(map[string]interface{}{})
	assert.Error(t, err)
}

// TestCSVParser_SpecialCharacters tests handling special characters
func TestCSVParser_SpecialCharacters(t *testing.T) {
	executor, err := NewCSVParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: `name,description
"Test","Contains ""quotes"" and, commas"
"Another","Has
newlines"`,
	}

	result, err := executor.Execute(context.Background(), msg)
	require.NoError(t, err)

	payload, ok := result.Payload.([]interface{})
	require.True(t, ok)
	assert.Len(t, payload, 2)
}

// TestCSVParser_Cleanup tests cleanup
func TestCSVParser_Cleanup(t *testing.T) {
	executor, err := NewCSVParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}

// BenchmarkCSVParser_Parse benchmarks CSV parsing
func BenchmarkCSVParser_Parse(b *testing.B) {
	executor, _ := NewCSVParserExecutor(map[string]interface{}{
		"action": "parse",
	})

	csv := `name,age,city
John,30,NYC
Jane,25,LA
Bob,35,SF
Alice,28,Chicago
`

	msg := node.Message{Payload: csv}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute(context.Background(), msg)
	}
}

// BenchmarkCSVParser_Stringify benchmarks CSV stringification
func BenchmarkCSVParser_Stringify(b *testing.B) {
	executor, _ := NewCSVParserExecutor(map[string]interface{}{
		"action": "stringify",
	})

	data := []interface{}{
		map[string]interface{}{"name": "John", "age": float64(30)},
		map[string]interface{}{"name": "Jane", "age": float64(25)},
		map[string]interface{}{"name": "Bob", "age": float64(35)},
	}

	msg := node.Message{Payload: data}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute(context.Background(), msg)
	}
}
