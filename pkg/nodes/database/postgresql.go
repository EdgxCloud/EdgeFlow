package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	_ "github.com/lib/pq"
)

// PostgreSQLConfig configuration for the PostgreSQL node
type PostgreSQLConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
	SSLMode  string `json:"sslMode"` // disable, require, verify-ca, verify-full
}

// PostgreSQLExecutor executor for the PostgreSQL node
type PostgreSQLExecutor struct {
	config PostgreSQLConfig
	db     *sql.DB
	mu     sync.RWMutex
}

// NewPostgreSQLExecutor creates a new PostgreSQLExecutor
func NewPostgreSQLExecutor() node.Executor {
	return &PostgreSQLExecutor{}
}

// Init initializes the PostgreSQL node with configuration
func (e *PostgreSQLExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var pgConfig PostgreSQLConfig
	if err := json.Unmarshal(configJSON, &pgConfig); err != nil {
		return fmt.Errorf("invalid postgresql config: %w", err)
	}

	// Default values
	if pgConfig.Port == 0 {
		pgConfig.Port = 5432
	}
	if pgConfig.SSLMode == "" {
		pgConfig.SSLMode = "disable"
	}

	// Validate
	if pgConfig.Host == "" {
		return fmt.Errorf("host is required")
	}
	if pgConfig.Database == "" {
		return fmt.Errorf("database is required")
	}
	if pgConfig.Username == "" {
		return fmt.Errorf("username is required")
	}

	e.config = pgConfig

	// Initialize connection
	if err := e.connect(); err != nil {
		return err
	}

	return nil
}

// connect connects to PostgreSQL
func (e *PostgreSQLExecutor) connect() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Build DSN
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		e.config.Host,
		e.config.Port,
		e.config.Username,
		e.config.Password,
		e.config.Database,
		e.config.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	e.db = db
	return nil
}

// Execute executes the node
func (e *PostgreSQLExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	e.mu.RLock()
	db := e.db
	e.mu.RUnlock()

	if db == nil {
		return node.Message{}, fmt.Errorf("not connected to database")
	}

	var query string
	var params []interface{}

	// Get query from message
	if q, ok := msg.Payload["query"].(string); ok {
		query = q
	}
	if p, ok := msg.Payload["params"].([]interface{}); ok {
		params = p
	}

	if query == "" {
		return node.Message{}, fmt.Errorf("query is required")
	}

	// Determine if it's a SELECT query
	isSelect := len(query) >= 6 && query[:6] == "SELECT"

	if isSelect {
		// Execute SELECT query
		rows, err := db.QueryContext(ctx, query, params...)
		if err != nil {
			return node.Message{}, fmt.Errorf("query failed: %w", err)
		}
		defer rows.Close()

		// Get column names
		columns, err := rows.Columns()
		if err != nil {
			return node.Message{}, err
		}

		// Read all rows
		var results []map[string]interface{}
		for rows.Next() {
			// Create a slice of interface{}'s to represent each column
			values := make([]interface{}, len(columns))
			valuePtrs := make([]interface{}, len(columns))
			for i := range values {
				valuePtrs[i] = &values[i]
			}

			// Scan the row
			if err := rows.Scan(valuePtrs...); err != nil {
				return node.Message{}, err
			}

			// Create map for this row
			row := make(map[string]interface{})
			for i, col := range columns {
				val := values[i]
				// Convert []byte to string
				if b, ok := val.([]byte); ok {
					row[col] = string(b)
				} else {
					row[col] = val
				}
			}
			results = append(results, row)
		}

		if err := rows.Err(); err != nil {
			return node.Message{}, err
		}

		return node.Message{
			Payload: map[string]interface{}{
				"results": results,
				"count":   len(results),
			},
		}, nil
	} else {
		// Execute INSERT/UPDATE/DELETE query
		result, err := db.ExecContext(ctx, query, params...)
		if err != nil {
			return node.Message{}, fmt.Errorf("query failed: %w", err)
		}

		rowsAffected, _ := result.RowsAffected()
		lastInsertId, _ := result.LastInsertId()

		return node.Message{
			Payload: map[string]interface{}{
				"rowsAffected": rowsAffected,
				"lastInsertId": lastInsertId,
			},
		}, nil
	}
}

// Cleanup releases resources
func (e *PostgreSQLExecutor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.db != nil {
		return e.db.Close()
	}
	return nil
}
