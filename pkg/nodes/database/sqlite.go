package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteNode executes SQL queries on SQLite databases
type SQLiteNode struct {
	dbPath     string
	query      string
	db         *sql.DB
	params     []interface{}
	outputKey  string
	autoCommit bool
}

// NewSQLiteNode creates a new SQLite node
func NewSQLiteNode() *SQLiteNode {
	return &SQLiteNode{
		outputKey:  "result",
		autoCommit: true,
	}
}

// Init initializes the SQLite node with configuration
func (n *SQLiteNode) Init(config map[string]interface{}) error {
	// Parse database path
	if dbPath, ok := config["database"].(string); ok {
		n.dbPath = dbPath
	} else {
		return fmt.Errorf("database path is required")
	}

	// Parse query
	if query, ok := config["query"].(string); ok {
		n.query = query
	}

	// Parse output key
	if outputKey, ok := config["outputKey"].(string); ok {
		n.outputKey = outputKey
	}

	// Parse auto commit
	if autoCommit, ok := config["autoCommit"].(bool); ok {
		n.autoCommit = autoCommit
	}

	// Open database connection
	db, err := sql.Open("sqlite3", n.dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	n.db = db

	// Run initialization queries if provided
	if initQueries, ok := config["initQueries"].([]interface{}); ok {
		for _, q := range initQueries {
			if queryStr, ok := q.(string); ok {
				if _, err := db.Exec(queryStr); err != nil {
					return fmt.Errorf("failed to execute init query: %w", err)
				}
			}
		}
	}

	return nil
}

// Execute processes incoming messages
func (n *SQLiteNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if n.db == nil {
		return node.Message{}, fmt.Errorf("database not initialized")
	}

	// Get query from message or use default
	query := n.query
	if msgQuery, ok := msg.Payload["query"].(string); ok && msgQuery != "" {
		query = msgQuery
	}

	if query == "" {
		return node.Message{}, fmt.Errorf("no query specified")
	}

	// Get parameters from message
	params := make([]interface{}, 0)
	if msgParams, ok := msg.Payload["params"].([]interface{}); ok {
		params = msgParams
	}

	// Determine query type
	queryType := strings.ToUpper(strings.TrimSpace(strings.Split(query, " ")[0]))

	var result interface{}
	var err error

	switch queryType {
	case "SELECT", "PRAGMA":
		result, err = n.executeQuery(ctx, query, params)
	case "INSERT", "UPDATE", "DELETE", "CREATE", "DROP", "ALTER":
		result, err = n.executeExec(ctx, query, params)
	default:
		return node.Message{}, fmt.Errorf("unsupported query type: %s", queryType)
	}

	if err != nil {
		return node.Message{}, err
	}

	// Create output message
	outputPayload := make(map[string]interface{})
	for k, v := range msg.Payload {
		outputPayload[k] = v
	}
	outputPayload[n.outputKey] = result
	outputPayload["queryType"] = queryType

	return node.Message{
		Type:    node.MessageTypeData,
		Payload: outputPayload,
		Topic:   msg.Topic,
	}, nil
}

// executeQuery executes a SELECT query and returns results
func (n *SQLiteNode) executeQuery(ctx context.Context, query string, params []interface{}) ([]map[string]interface{}, error) {
	rows, err := n.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	// Prepare result slice
	results := make([]map[string]interface{}, 0)

	// Iterate through rows
	for rows.Next() {
		// Create a slice of interface{} to hold the values
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		// Scan the row
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		// Create a map for this row
		rowMap := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]

			// Convert []byte to string
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}

		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// executeExec executes an INSERT, UPDATE, DELETE or DDL statement
func (n *SQLiteNode) executeExec(ctx context.Context, query string, params []interface{}) (map[string]interface{}, error) {
	result, err := n.db.ExecContext(ctx, query, params...)
	if err != nil {
		return nil, fmt.Errorf("exec failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		rowsAffected = 0
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		lastInsertId = 0
	}

	return map[string]interface{}{
		"rowsAffected": rowsAffected,
		"lastInsertId": lastInsertId,
	}, nil
}

// Cleanup closes the database connection
func (n *SQLiteNode) Cleanup() error {
	if n.db != nil {
		return n.db.Close()
	}
	return nil
}

// BeginTransaction starts a new transaction
func (n *SQLiteNode) BeginTransaction(ctx context.Context) (*sql.Tx, error) {
	if n.db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return n.db.BeginTx(ctx, nil)
}
