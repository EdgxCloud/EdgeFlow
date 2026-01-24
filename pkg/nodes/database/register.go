package database

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterAllNodes ثبت تمام نودهای database
func RegisterAllNodes(registry *node.Registry) {
	// MySQL
	registry.Register(&node.NodeInfo{
		Type:        "mysql",
		Name:        "MySQL",
		Category:    node.NodeTypeProcessing,
		Description: "اجرای query در MySQL",
		Icon:        "database",
		Properties: []node.PropertySchema{
			{Name: "host", Label: "Host", Type: "string", Default: "localhost", Required: true},
			{Name: "port", Label: "Port", Type: "number", Default: 3306, Required: true},
			{Name: "database", Label: "Database", Type: "string", Default: "", Required: true},
			{Name: "username", Label: "Username", Type: "string", Default: "", Required: true},
			{Name: "password", Label: "Password", Type: "string", Default: "", Required: false},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: NewMySQLExecutor,
	})

	// Redis
	registry.Register(&node.NodeInfo{
		Type:        "redis",
		Name:        "Redis",
		Category:    node.NodeTypeProcessing,
		Description: "عملیات Redis (get, set, delete, incr, decr)",
		Icon:        "box",
		Properties: []node.PropertySchema{
			{Name: "host", Label: "Host", Type: "string", Default: "localhost", Required: true},
			{Name: "port", Label: "Port", Type: "number", Default: 6379, Required: true},
			{Name: "password", Label: "Password", Type: "string", Default: "", Required: false},
			{Name: "db", Label: "Database", Type: "number", Default: 0, Required: false},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: NewRedisExecutor,
	})

	// PostgreSQL
	registry.Register(&node.NodeInfo{
		Type:        "postgresql",
		Name:        "PostgreSQL",
		Category:    node.NodeTypeProcessing,
		Description: "اجرای query در PostgreSQL",
		Icon:        "database",
		Properties: []node.PropertySchema{
			{Name: "host", Label: "Host", Type: "string", Default: "localhost", Required: true},
			{Name: "port", Label: "Port", Type: "number", Default: 5432, Required: true},
			{Name: "database", Label: "Database", Type: "string", Default: "", Required: true},
			{Name: "username", Label: "Username", Type: "string", Default: "", Required: true},
			{Name: "password", Label: "Password", Type: "string", Default: "", Required: false},
			{Name: "sslMode", Label: "SSL Mode", Type: "select", Default: "disable", Required: false, Options: []string{"disable", "require", "verify-ca", "verify-full"}},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: NewPostgreSQLExecutor,
	})

	// MongoDB
	registry.Register(&node.NodeInfo{
		Type:        "mongodb",
		Name:        "MongoDB",
		Category:    node.NodeTypeProcessing,
		Description: "عملیات MongoDB (find, insert, update, delete)",
		Icon:        "database",
		Properties: []node.PropertySchema{
			{Name: "uri", Label: "URI", Type: "string", Default: "mongodb://localhost:27017", Required: true},
			{Name: "database", Label: "Database", Type: "string", Default: "", Required: true},
			{Name: "collection", Label: "Collection", Type: "string", Default: "", Required: false},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: NewMongoDBExecutor,
	})

	// SQLite
	registry.Register(&node.NodeInfo{
		Type:        "sqlite",
		Name:        "SQLite",
		Category:    node.NodeTypeProcessing,
		Description: "Execute SQL queries on SQLite database",
		Icon:        "database",
		Properties: []node.PropertySchema{
			{Name: "database", Label: "Database", Type: "string", Default: "data.db", Required: true},
			{Name: "query", Label: "Query", Type: "string", Default: "", Required: false},
			{Name: "outputKey", Label: "Output Key", Type: "string", Default: "result", Required: false},
			{Name: "autoCommit", Label: "Auto Commit", Type: "boolean", Default: true, Required: false},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewSQLiteNode()
		},
	})

	// InfluxDB
	registry.Register(&node.NodeInfo{
		Type:        "influxdb",
		Name:        "InfluxDB",
		Category:    node.NodeTypeProcessing,
		Description: "Time-series database operations (write, query, delete)",
		Icon:        "database",
		Properties: []node.PropertySchema{
			{Name: "url", Label: "URL", Type: "string", Default: "http://localhost:8086", Required: true},
			{Name: "token", Label: "Token", Type: "string", Default: "", Required: true},
			{Name: "org", Label: "Organization", Type: "string", Default: "", Required: true},
			{Name: "bucket", Label: "Bucket", Type: "string", Default: "", Required: true},
			{Name: "measurement", Label: "Measurement", Type: "string", Default: "", Required: false},
			{Name: "precision", Label: "Precision", Type: "select", Default: "ns", Required: false, Options: []string{"ns", "us", "ms", "s"}},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewInfluxDBNode()
		},
	})
}
