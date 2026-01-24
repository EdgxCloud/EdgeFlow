package database

import (
	"context"
	"fmt"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
)

// InfluxDBNode handles time-series data operations with InfluxDB
type InfluxDBNode struct {
	url           string
	token         string
	org           string
	bucket        string
	measurement   string
	client        influxdb2.Client
	writeAPI      api.WriteAPIBlocking
	queryAPI      api.QueryAPI
	tags          map[string]string
	precision     time.Duration
	batchSize     int
	flushInterval time.Duration
}

// NewInfluxDBNode creates a new InfluxDB node
func NewInfluxDBNode() *InfluxDBNode {
	return &InfluxDBNode{
		precision:     time.Nanosecond,
		batchSize:     100,
		flushInterval: 1 * time.Second,
		tags:          make(map[string]string),
	}
}

// Init initializes the InfluxDB node with configuration
func (n *InfluxDBNode) Init(config map[string]interface{}) error {
	// Parse URL
	if url, ok := config["url"].(string); ok {
		n.url = url
	} else {
		return fmt.Errorf("url is required")
	}

	// Parse token
	if token, ok := config["token"].(string); ok {
		n.token = token
	} else {
		return fmt.Errorf("token is required")
	}

	// Parse organization
	if org, ok := config["org"].(string); ok {
		n.org = org
	} else {
		return fmt.Errorf("organization is required")
	}

	// Parse bucket
	if bucket, ok := config["bucket"].(string); ok {
		n.bucket = bucket
	} else {
		return fmt.Errorf("bucket is required")
	}

	// Parse measurement (optional for write operations)
	if measurement, ok := config["measurement"].(string); ok {
		n.measurement = measurement
	}

	// Parse tags (optional)
	if tags, ok := config["tags"].(map[string]interface{}); ok {
		for k, v := range tags {
			if strVal, ok := v.(string); ok {
				n.tags[k] = strVal
			}
		}
	}

	// Parse precision
	if precisionStr, ok := config["precision"].(string); ok {
		switch precisionStr {
		case "ns":
			n.precision = time.Nanosecond
		case "us":
			n.precision = time.Microsecond
		case "ms":
			n.precision = time.Millisecond
		case "s":
			n.precision = time.Second
		default:
			return fmt.Errorf("invalid precision: %s (use: ns, us, ms, s)", precisionStr)
		}
	}

	// Parse batch size
	if batchSize, ok := config["batchSize"].(float64); ok {
		n.batchSize = int(batchSize)
	}

	// Parse flush interval
	if flushIntervalStr, ok := config["flushInterval"].(string); ok {
		duration, err := time.ParseDuration(flushIntervalStr)
		if err != nil {
			return fmt.Errorf("invalid flush interval: %w", err)
		}
		n.flushInterval = duration
	}

	// Create InfluxDB client
	n.client = influxdb2.NewClient(n.url, n.token)

	// Create write API
	n.writeAPI = n.client.WriteAPIBlocking(n.org, n.bucket)

	// Create query API
	n.queryAPI = n.client.QueryAPI(n.org)

	// Verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health, err := n.client.Health(ctx)
	if err != nil {
		n.client.Close()
		return fmt.Errorf("failed to connect to InfluxDB: %w", err)
	}

	if health.Status != "pass" {
		n.client.Close()
		return fmt.Errorf("InfluxDB health check failed: %s", health.Status)
	}

	return nil
}

// Execute processes incoming messages
func (n *InfluxDBNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if n.client == nil {
		return node.Message{}, fmt.Errorf("InfluxDB client not initialized")
	}

	// Determine operation type
	operation := "write"
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	var result interface{}
	var err error

	switch operation {
	case "write":
		result, err = n.writeData(ctx, msg)
	case "query":
		result, err = n.queryData(ctx, msg)
	case "delete":
		result, err = n.deleteData(ctx, msg)
	default:
		return node.Message{}, fmt.Errorf("unsupported operation: %s", operation)
	}

	if err != nil {
		return node.Message{}, err
	}

	// Create output message
	outputPayload := make(map[string]interface{})
	for k, v := range msg.Payload {
		outputPayload[k] = v
	}
	outputPayload["result"] = result
	outputPayload["operation"] = operation

	return node.Message{
		Type:    node.MessageTypeData,
		Payload: outputPayload,
		Topic:   msg.Topic,
	}, nil
}

// writeData writes data points to InfluxDB
func (n *InfluxDBNode) writeData(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get measurement from message or config
	measurement := n.measurement
	if msgMeasurement, ok := msg.Payload["measurement"].(string); ok {
		measurement = msgMeasurement
	}

	if measurement == "" {
		return nil, fmt.Errorf("measurement is required for write operation")
	}

	// Get tags
	tags := make(map[string]string)
	for k, v := range n.tags {
		tags[k] = v
	}
	if msgTags, ok := msg.Payload["tags"].(map[string]interface{}); ok {
		for k, v := range msgTags {
			if strVal, ok := v.(string); ok {
				tags[k] = strVal
			}
		}
	}

	// Get fields
	fields := make(map[string]interface{})
	if msgFields, ok := msg.Payload["fields"].(map[string]interface{}); ok {
		fields = msgFields
	} else {
		return nil, fmt.Errorf("fields are required for write operation")
	}

	// Get timestamp (optional)
	var timestamp time.Time
	if msgTimestamp, ok := msg.Payload["timestamp"].(int64); ok {
		timestamp = time.Unix(msgTimestamp, 0)
	} else {
		timestamp = time.Now()
	}

	// Create point
	point := write.NewPoint(measurement, tags, fields, timestamp)

	// Write point
	err := n.writeAPI.WritePoint(ctx, point)
	if err != nil {
		return nil, fmt.Errorf("failed to write point: %w", err)
	}

	return map[string]interface{}{
		"measurement": measurement,
		"tags":        tags,
		"fields":      fields,
		"timestamp":   timestamp.Unix(),
		"written":     true,
	}, nil
}

// queryData queries data from InfluxDB using Flux
func (n *InfluxDBNode) queryData(ctx context.Context, msg node.Message) ([]map[string]interface{}, error) {
	// Get query from message
	query, ok := msg.Payload["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required for query operation")
	}

	// Execute query
	result, err := n.queryAPI.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer result.Close()

	// Parse results
	results := make([]map[string]interface{}, 0)

	for result.Next() {
		record := result.Record()

		row := make(map[string]interface{})
		row["time"] = record.Time().Unix()
		row["measurement"] = record.Measurement()

		// Add fields
		for k, v := range record.Values() {
			row[k] = v
		}

		results = append(results, row)
	}

	if result.Err() != nil {
		return nil, fmt.Errorf("error reading query results: %w", result.Err())
	}

	return results, nil
}

// deleteData deletes data from InfluxDB
func (n *InfluxDBNode) deleteData(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get time range as strings
	startStr, ok := msg.Payload["start"].(string)
	if !ok {
		return nil, fmt.Errorf("start time is required for delete operation")
	}

	stopStr, ok := msg.Payload["stop"].(string)
	if !ok {
		return nil, fmt.Errorf("stop time is required for delete operation")
	}

	// Parse start time
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start time format (expected RFC3339): %w", err)
	}

	// Parse stop time
	stop, err := time.Parse(time.RFC3339, stopStr)
	if err != nil {
		return nil, fmt.Errorf("invalid stop time format (expected RFC3339): %w", err)
	}

	// Get predicate (optional)
	predicate := ""
	if pred, ok := msg.Payload["predicate"].(string); ok {
		predicate = pred
	}

	// Delete data
	err = n.client.DeleteAPI().DeleteWithName(ctx, n.org, n.bucket, start, stop, predicate)
	if err != nil {
		return nil, fmt.Errorf("failed to delete data: %w", err)
	}

	return map[string]interface{}{
		"deleted": true,
		"start":   startStr,
		"stop":    stopStr,
	}, nil
}

// Cleanup closes the InfluxDB connection
func (n *InfluxDBNode) Cleanup() error {
	if n.client != nil {
		n.client.Close()
	}
	return nil
}
