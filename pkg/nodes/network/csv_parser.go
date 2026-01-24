package network

import (
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	"github.com/edgeflow/edgeflow/internal/node"
)

type CSVParserNode struct {
	action    string
	delimiter rune
	hasHeader bool
	skipRows  int
}

func NewCSVParserNode() *CSVParserNode {
	return &CSVParserNode{
		action:    "parse",
		delimiter: ',',
		hasHeader: true,
		skipRows:  0,
	}
}

func (n *CSVParserNode) Init(config map[string]interface{}) error {
	if action, ok := config["action"].(string); ok {
		n.action = action
	}
	if delim, ok := config["delimiter"].(string); ok && len(delim) > 0 {
		n.delimiter = rune(delim[0])
	}
	if hasHeader, ok := config["hasHeader"].(bool); ok {
		n.hasHeader = hasHeader
	}
	if skip, ok := config["skipRows"].(float64); ok {
		n.skipRows = int(skip)
	} else if skip, ok := config["skipRows"].(int); ok {
		n.skipRows = skip
	}
	return nil
}

func (n *CSVParserNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	switch n.action {
	case "parse":
		return n.parseCSV(msg)
	case "stringify":
		return n.stringifyCSV(msg)
	default:
		return msg, fmt.Errorf("unknown action: %s", n.action)
	}
}

func (n *CSVParserNode) parseCSV(msg node.Message) (node.Message, error) {
	var csvStr string

	// Handle both map[string]interface{} and direct string/[]byte
	payload := msg.Payload
	if str, ok := payload["data"].(string); ok {
		csvStr = str
	} else if bytes, ok := payload["data"].([]byte); ok {
		csvStr = string(bytes)
	} else {
		return msg, fmt.Errorf("payload must contain 'data' as string or []byte")
	}

	reader := csv.NewReader(strings.NewReader(csvStr))
	reader.Comma = n.delimiter

	records, err := reader.ReadAll()
	if err != nil {
		return msg, fmt.Errorf("failed to parse CSV: %w", err)
	}

	if n.skipRows > 0 && len(records) > n.skipRows {
		records = records[n.skipRows:]
	}

	if len(records) == 0 {
		msg.Payload = map[string]interface{}{"data": []interface{}{}}
		return msg, nil
	}

	var result []interface{}

	if n.hasHeader && len(records) > 1 {
		headers := records[0]
		for _, record := range records[1:] {
			row := make(map[string]interface{})
			for i, value := range record {
				if i < len(headers) {
					row[headers[i]] = value
				}
			}
			result = append(result, row)
		}
	} else {
		for _, record := range records {
			row := make([]interface{}, len(record))
			for i, value := range record {
				row[i] = value
			}
			result = append(result, row)
		}
	}

	msg.Payload = map[string]interface{}{"data": result}
	return msg, nil
}

func (n *CSVParserNode) stringifyCSV(msg node.Message) (node.Message, error) {
	var records [][]string
	var dataToStringify interface{}

	// Extract data from payload if it's a map
	payload := msg.Payload
	if data, ok := payload["data"]; ok {
		dataToStringify = data
	} else {
		dataToStringify = msg.Payload
	}

	switch v := dataToStringify.(type) {
	case []interface{}:
		for _, item := range v {
			if rowMap, ok := item.(map[string]interface{}); ok {
				var row []string
				for _, value := range rowMap {
					row = append(row, fmt.Sprintf("%v", value))
				}
				records = append(records, row)
			} else if rowArr, ok := item.([]interface{}); ok {
				var row []string
				for _, value := range rowArr {
					row = append(row, fmt.Sprintf("%v", value))
				}
				records = append(records, row)
			}
		}
	default:
		return msg, fmt.Errorf("payload must be array")
	}

	var builder strings.Builder
	writer := csv.NewWriter(&builder)
	writer.Comma = n.delimiter

	if err := writer.WriteAll(records); err != nil {
		return msg, fmt.Errorf("failed to stringify CSV: %w", err)
	}

	msg.Payload = map[string]interface{}{"data": builder.String()}
	return msg, nil
}

func (n *CSVParserNode) Cleanup() error {
	return nil
}

func NewCSVParserExecutor() node.Executor {
	return NewCSVParserNode()
}
