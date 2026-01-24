package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage implements Storage using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite-based storage
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteStorage{db: db}

	if err := storage.init(); err != nil {
		db.Close()
		return nil, err
	}

	return storage, nil
}

// init creates the necessary tables
func (s *SQLiteStorage) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS flows (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT,
		status TEXT,
		data TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_flows_name ON flows(name);
	CREATE INDEX IF NOT EXISTS idx_flows_status ON flows(status);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// SaveFlow saves a flow to the database
func (s *SQLiteStorage) SaveFlow(flow *Flow) error {
	data, err := json.Marshal(flow)
	if err != nil {
		return fmt.Errorf("failed to marshal flow: %w", err)
	}

	query := `
		INSERT INTO flows (id, name, description, status, data)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			status = excluded.status,
			data = excluded.data,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = s.db.Exec(query, flow.ID, flow.Name, flow.Description, flow.Status, string(data))
	if err != nil {
		return fmt.Errorf("failed to save flow: %w", err)
	}

	return nil
}

// GetFlow retrieves a flow from the database
func (s *SQLiteStorage) GetFlow(id string) (*Flow, error) {
	query := `SELECT data FROM flows WHERE id = ?`

	var data string
	err := s.db.QueryRow(query, id).Scan(&data)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("flow not found: %s", id)
		}
		return nil, fmt.Errorf("failed to query flow: %w", err)
	}

	var flow Flow
	if err := json.Unmarshal([]byte(data), &flow); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flow: %w", err)
	}

	return &flow, nil
}

// ListFlows returns all flows from the database
func (s *SQLiteStorage) ListFlows() ([]*Flow, error) {
	query := `SELECT data FROM flows ORDER BY updated_at DESC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query flows: %w", err)
	}
	defer rows.Close()

	flows := []*Flow{}

	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			continue
		}

		var flow Flow
		if err := json.Unmarshal([]byte(data), &flow); err != nil {
			continue
		}

		flows = append(flows, &flow)
	}

	return flows, nil
}

// DeleteFlow removes a flow from the database
func (s *SQLiteStorage) DeleteFlow(id string) error {
	query := `DELETE FROM flows WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete flow: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("flow not found: %s", id)
	}

	return nil
}

// UpdateFlow updates an existing flow in the database
func (s *SQLiteStorage) UpdateFlow(flow *Flow) error {
	// For SQLite, update is the same as save (using UPSERT)
	return s.SaveFlow(flow)
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
