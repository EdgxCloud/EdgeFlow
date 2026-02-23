package database

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteNode_Init(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"database": dbPath,
				"query":    "SELECT * FROM users",
			},
			wantErr: false,
		},
		{
			name: "missing database path",
			config: map[string]interface{}{
				"query": "SELECT * FROM users",
			},
			wantErr: true,
		},
		{
			name: "with init queries",
			config: map[string]interface{}{
				"database": filepath.Join(tmpDir, "test2.db"),
				"initQueries": []interface{}{
					"CREATE TABLE IF NOT EXISTS test (id INTEGER PRIMARY KEY, name TEXT)",
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewSQLiteNode()
			err := n.Init(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				n.Cleanup()
			}
		})
	}
}

func TestSQLiteNode_Execute_SELECT(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	n := NewSQLiteNode()
	err := n.Init(map[string]interface{}{
		"database": dbPath,
		"initQueries": []interface{}{
			"CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, email TEXT)",
			"INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')",
			"INSERT INTO users (name, email) VALUES ('Jane Smith', 'jane@example.com')",
		},
	})
	require.NoError(t, err)
	defer n.Cleanup()

	ctx := context.Background()

	// Test SELECT query
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"query": "SELECT * FROM users",
		},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	assert.Equal(t, node.MessageTypeData, result.Type)

	// Check result
	resultData, ok := result.Payload["result"].([]map[string]interface{})
	require.True(t, ok, "Result should be a slice of maps")
	assert.Len(t, resultData, 2)
	assert.Equal(t, "John Doe", resultData[0]["name"])
	assert.Equal(t, "jane@example.com", resultData[1]["email"])
}

func TestSQLiteNode_Execute_INSERT(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	n := NewSQLiteNode()
	err := n.Init(map[string]interface{}{
		"database": dbPath,
		"initQueries": []interface{}{
			"CREATE TABLE logs (id INTEGER PRIMARY KEY AUTOINCREMENT, message TEXT, timestamp INTEGER)",
		},
	})
	require.NoError(t, err)
	defer n.Cleanup()

	ctx := context.Background()

	// Test INSERT query
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"query":  "INSERT INTO logs (message, timestamp) VALUES (?, ?)",
			"params": []interface{}{"Test message", 1234567890},
		},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	assert.Equal(t, node.MessageTypeData, result.Type)

	// Check result
	resultData, ok := result.Payload["result"].(map[string]interface{})
	require.True(t, ok, "Result should be a map")
	assert.Equal(t, int64(1), resultData["rowsAffected"])
	assert.Equal(t, int64(1), resultData["lastInsertId"])
}

func TestSQLiteNode_Execute_UPDATE(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	n := NewSQLiteNode()
	err := n.Init(map[string]interface{}{
		"database": dbPath,
		"initQueries": []interface{}{
			"CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT)",
			"INSERT INTO settings (key, value) VALUES ('theme', 'dark')",
			"INSERT INTO settings (key, value) VALUES ('language', 'en')",
		},
	})
	require.NoError(t, err)
	defer n.Cleanup()

	ctx := context.Background()

	// Test UPDATE query
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"query":  "UPDATE settings SET value = ? WHERE key = ?",
			"params": []interface{}{"light", "theme"},
		},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	// Check result
	resultData, ok := result.Payload["result"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(1), resultData["rowsAffected"])
}

func TestSQLiteNode_Execute_DELETE(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	n := NewSQLiteNode()
	err := n.Init(map[string]interface{}{
		"database": dbPath,
		"initQueries": []interface{}{
			"CREATE TABLE temp_data (id INTEGER PRIMARY KEY, data TEXT)",
			"INSERT INTO temp_data (data) VALUES ('test1')",
			"INSERT INTO temp_data (data) VALUES ('test2')",
		},
	})
	require.NoError(t, err)
	defer n.Cleanup()

	ctx := context.Background()

	// Test DELETE query
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"query":  "DELETE FROM temp_data WHERE id = ?",
			"params": []interface{}{1},
		},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	// Check result
	resultData, ok := result.Payload["result"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int64(1), resultData["rowsAffected"])
}

func TestSQLiteNode_Execute_WithDefaultQuery(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	n := NewSQLiteNode()
	err := n.Init(map[string]interface{}{
		"database": dbPath,
		"query":    "SELECT 1 as value",
	})
	require.NoError(t, err)
	defer n.Cleanup()

	ctx := context.Background()

	// Execute without query in message (should use default)
	msg := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	resultData, ok := result.Payload["result"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, resultData, 1)
	assert.Equal(t, int64(1), resultData[0]["value"])
}

func TestSQLiteNode_Execute_Error(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	n := NewSQLiteNode()
	err := n.Init(map[string]interface{}{
		"database": dbPath,
	})
	require.NoError(t, err)
	defer n.Cleanup()

	ctx := context.Background()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "invalid SQL syntax",
			query:   "SELEKT * FROM users",
			wantErr: true,
		},
		{
			name:    "table not found",
			query:   "SELECT * FROM nonexistent_table",
			wantErr: true,
		},
		{
			name:    "empty query",
			query:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{
				Type: node.MessageTypeData,
				Payload: map[string]interface{}{
					"query": tt.query,
				},
			}

			_, err := n.Execute(ctx, msg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSQLiteNode_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	n := NewSQLiteNode()
	err := n.Init(map[string]interface{}{
		"database": dbPath,
	})
	require.NoError(t, err)

	err = n.Cleanup()
	assert.NoError(t, err)

	// Verify database file was created
	_, err = os.Stat(dbPath)
	assert.NoError(t, err, "Database file should exist")
}

func TestSQLiteNode_Transaction(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	n := NewSQLiteNode()
	err := n.Init(map[string]interface{}{
		"database": dbPath,
		"initQueries": []interface{}{
			"CREATE TABLE accounts (id INTEGER PRIMARY KEY, balance REAL)",
			"INSERT INTO accounts (balance) VALUES (100.0)",
		},
	})
	require.NoError(t, err)
	defer n.Cleanup()

	ctx := context.Background()

	// Begin transaction
	tx, err := n.BeginTransaction(ctx)
	require.NoError(t, err)

	// Execute transaction queries
	_, err = tx.ExecContext(ctx, "UPDATE accounts SET balance = balance - 50 WHERE id = 1")
	require.NoError(t, err)

	// Commit transaction
	err = tx.Commit()
	require.NoError(t, err)

	// Verify result
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"query": "SELECT balance FROM accounts WHERE id = 1",
		},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	resultData := result.Payload["result"].([]map[string]interface{})
	assert.Equal(t, 50.0, resultData[0]["balance"])
}
