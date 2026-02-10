package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"

	"github.com/edgeflow/edgeflow/internal/node"
)

// MySQLConfig configuration for the MySQL node
type MySQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// MySQLExecutor executor for the MySQL node
type MySQLExecutor struct {
	config MySQLConfig
	db     *sql.DB
	mu     sync.RWMutex
}

// NewMySQLExecutor creates a new MySQLExecutor
func NewMySQLExecutor() node.Executor {
	return &MySQLExecutor{}
}

// Init initializes the MySQL node with configuration
func (e *MySQLExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var mysqlConfig MySQLConfig
	if err := json.Unmarshal(configJSON, &mysqlConfig); err != nil {
		return fmt.Errorf("invalid mysql config: %w", err)
	}

	// Validate
	if mysqlConfig.Host == "" {
		return fmt.Errorf("host is required")
	}
	if mysqlConfig.Port == 0 {
		mysqlConfig.Port = 3306
	}
	if mysqlConfig.Database == "" {
		return fmt.Errorf("database is required")
	}

	e.config = mysqlConfig
	return nil
}

// Execute executes the node
func (e *MySQLExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Connect if not connected
	if e.db == nil {
		if err := e.connect(); err != nil {
			return node.Message{}, fmt.Errorf("failed to connect: %w", err)
		}
	}

	var query string
	var params []interface{}

	// Get query from message
	if q, ok := msg.Payload["query"].(string); ok {
		query = q
	} else if q, ok := msg.Payload["sql"].(string); ok {
		query = q
	}

	if p, ok := msg.Payload["params"].([]interface{}); ok {
		params = p
	}

	if query == "" {
		return node.Message{}, fmt.Errorf("query is required")
	}

	// Execute query
	rows, err := e.db.QueryContext(ctx, query, params...)
	if err != nil {
		return node.Message{}, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return node.Message{}, err
	}

	// Fetch results
	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return node.Message{}, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			var v interface{}
			val := values[i]
			b, ok := val.([]byte)
			if ok {
				v = string(b)
			} else {
				v = val
			}
			row[col] = v
		}
		results = append(results, row)
	}

	return node.Message{
		Payload: map[string]interface{}{
			"results": results,
			"count":   len(results),
			"columns": columns,
		},
	}, nil
}

// connect connects to MySQL
func (e *MySQLExecutor) connect() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.db != nil {
		return nil
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		e.config.Username,
		e.config.Password,
		e.config.Host,
		e.config.Port,
		e.config.Database,
	)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return err
	}

	e.db = db
	return nil
}

// Cleanup releases resources
func (e *MySQLExecutor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.db != nil {
		e.db.Close()
		e.db = nil
	}
	return nil
}
