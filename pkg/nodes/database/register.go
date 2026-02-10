package database

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterAllNodes ثبت تمام نودهای database
func RegisterAllNodes(registry *node.Registry) {
	// ============================================
	// SQL DATABASES (3 nodes)
	// ============================================

	// MySQL
	registry.Register(&node.NodeInfo{
		Type:        "mysql",
		Name:        "MySQL",
		Category:    node.NodeTypeProcessing,
		Description: "اجرای query در MySQL",
		Icon:        "database",
		Color:       "#00758f",
		Properties: []node.PropertySchema{
			{Name: "host", Label: "Host", Type: "string", Default: "localhost", Required: true, Description: "MySQL server hostname or IP"},
			{Name: "port", Label: "Port", Type: "number", Default: 3306, Required: true, Description: "MySQL server port"},
			{Name: "database", Label: "Database", Type: "string", Default: "", Required: true, Description: "Database name to connect to"},
			{Name: "username", Label: "Username", Type: "string", Default: "", Required: true, Description: "Database username"},
			{Name: "password", Label: "Password", Type: "string", Default: "", Description: "Database password"},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Query", Type: "any", Description: "SQL query or operation to execute"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Result", Type: "any", Description: "Query result rows or affected count"},
		},
		Factory: NewMySQLExecutor,
	})

	// PostgreSQL
	registry.Register(&node.NodeInfo{
		Type:        "postgresql",
		Name:        "PostgreSQL",
		Category:    node.NodeTypeProcessing,
		Description: "اجرای query در PostgreSQL",
		Icon:        "database",
		Color:       "#336791",
		Properties: []node.PropertySchema{
			{Name: "host", Label: "Host", Type: "string", Default: "localhost", Required: true, Description: "PostgreSQL server hostname or IP"},
			{Name: "port", Label: "Port", Type: "number", Default: 5432, Required: true, Description: "PostgreSQL server port"},
			{Name: "database", Label: "Database", Type: "string", Default: "", Required: true, Description: "Database name to connect to"},
			{Name: "username", Label: "Username", Type: "string", Default: "", Required: true, Description: "Database username"},
			{Name: "password", Label: "Password", Type: "string", Default: "", Description: "Database password"},
			{Name: "sslMode", Label: "SSL Mode", Type: "select", Default: "disable", Description: "SSL connection mode", Options: []string{"disable", "require", "verify-ca", "verify-full"}},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Query", Type: "any", Description: "SQL query or operation to execute"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Result", Type: "any", Description: "Query result rows or affected count"},
		},
		Factory: NewPostgreSQLExecutor,
	})

	// SQLite
	registry.Register(&node.NodeInfo{
		Type:        "sqlite",
		Name:        "SQLite",
		Category:    node.NodeTypeProcessing,
		Description: "اجرای query در SQLite",
		Icon:        "database",
		Color:       "#003b57",
		Properties: []node.PropertySchema{
			{Name: "database", Label: "Database File", Type: "string", Default: "data.db", Required: true, Description: "SQLite database file path"},
			{Name: "query", Label: "Query", Type: "string", Default: "", Description: "Default SQL query (can be set via msg.query)"},
			{Name: "outputKey", Label: "Output Key", Type: "string", Default: "result", Description: "Key name for result in output payload"},
			{Name: "autoCommit", Label: "Auto Commit", Type: "boolean", Default: true, Description: "Automatically commit after each query"},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Query", Type: "any", Description: "SQL query or operation to execute"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Result", Type: "any", Description: "Query result rows or affected count"},
		},
		Factory: func() node.Executor {
			return NewSQLiteNode()
		},
	})

	// ============================================
	// NoSQL DATABASES (2 nodes)
	// ============================================

	// MongoDB
	registry.Register(&node.NodeInfo{
		Type:        "mongodb",
		Name:        "MongoDB",
		Category:    node.NodeTypeProcessing,
		Description: "عملیات MongoDB (find, insert, update, delete)",
		Icon:        "database",
		Color:       "#47a248",
		Properties: []node.PropertySchema{
			{Name: "uri", Label: "Connection URI", Type: "string", Default: "mongodb://localhost:27017", Required: true, Description: "MongoDB connection string"},
			{Name: "database", Label: "Database", Type: "string", Default: "", Required: true, Description: "Database name"},
			{Name: "collection", Label: "Collection", Type: "string", Default: "", Description: "Default collection (can be set via msg.collection)"},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Operation", Type: "any", Description: "MongoDB operation (find, insert, update, delete)"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Result", Type: "any", Description: "Operation result"},
		},
		Factory: NewMongoDBExecutor,
	})

	// Redis
	registry.Register(&node.NodeInfo{
		Type:        "redis",
		Name:        "Redis",
		Category:    node.NodeTypeProcessing,
		Description: "عملیات Redis (get, set, delete, incr, decr)",
		Icon:        "box",
		Color:       "#dc382d",
		Properties: []node.PropertySchema{
			{Name: "host", Label: "Host", Type: "string", Default: "localhost", Required: true, Description: "Redis server hostname or IP"},
			{Name: "port", Label: "Port", Type: "number", Default: 6379, Required: true, Description: "Redis server port"},
			{Name: "password", Label: "Password", Type: "string", Default: "", Description: "Redis server password"},
			{Name: "db", Label: "Database Index", Type: "number", Default: 0, Description: "Redis database index (0-15)"},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Command", Type: "any", Description: "Redis command (get, set, delete, incr, decr)"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Result", Type: "any", Description: "Command result"},
		},
		Factory: NewRedisExecutor,
	})

	// ============================================
	// TIME-SERIES DATABASES (1 node)
	// ============================================

	// InfluxDB
	registry.Register(&node.NodeInfo{
		Type:        "influxdb",
		Name:        "InfluxDB",
		Category:    node.NodeTypeProcessing,
		Description: "عملیات InfluxDB (write, query, delete)",
		Icon:        "database",
		Color:       "#22adf6",
		Properties: []node.PropertySchema{
			{Name: "url", Label: "URL", Type: "string", Default: "http://localhost:8086", Required: true, Description: "InfluxDB server URL"},
			{Name: "token", Label: "API Token", Type: "string", Default: "", Required: true, Description: "InfluxDB authentication token"},
			{Name: "org", Label: "Organization", Type: "string", Default: "", Required: true, Description: "InfluxDB organization name"},
			{Name: "bucket", Label: "Bucket", Type: "string", Default: "", Required: true, Description: "Target bucket for read/write operations"},
			{Name: "measurement", Label: "Measurement", Type: "string", Default: "", Description: "Default measurement name (can be set via msg)"},
			{Name: "precision", Label: "Write Precision", Type: "select", Default: "ns", Description: "Timestamp precision for writes", Options: []string{"ns", "us", "ms", "s"}},
			{Name: "batchSize", Label: "Batch Size", Type: "number", Default: 1000, Description: "Number of points to batch before flushing"},
			{Name: "flushInterval", Label: "Flush Interval", Type: "string", Default: "1s", Description: "Automatic flush interval (e.g. 1s, 500ms)"},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Data", Type: "any", Description: "Data point(s) to write or Flux query to execute"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Result", Type: "any", Description: "Query results or write confirmation"},
		},
		Factory: func() node.Executor {
			return NewInfluxDBNode()
		},
	})
}
