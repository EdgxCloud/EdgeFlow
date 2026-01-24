package dashboard

import (
	"context"

	"github.com/edgeflow/edgeflow/internal/node"
)

// TableColumn defines a table column
type TableColumn struct {
	Field      string `json:"field"`
	Header     string `json:"header"`
	Width      string `json:"width,omitempty"`
	Sortable   bool   `json:"sortable"`
	Filterable bool   `json:"filterable"`
	Format     string `json:"format,omitempty"`
}

// TableNode displays data in a table
type TableNode struct {
	*BaseWidget
	columns       []TableColumn
	rows          []map[string]interface{}
	maxRows       int
	pagination    bool
	rowsPerPage   int
	searchable    bool
	exportable    bool
	striped       bool
	bordered      bool
	compact       bool
}

// NewTableNode creates a new table widget
func NewTableNode() *TableNode {
	return &TableNode{
		BaseWidget:  NewBaseWidget(WidgetTypeTable),
		columns:     make([]TableColumn, 0),
		rows:        make([]map[string]interface{}, 0),
		maxRows:     1000,
		pagination:  true,
		rowsPerPage: 10,
		searchable:  true,
		exportable:  true,
		striped:     true,
		bordered:    false,
		compact:     false,
	}
}

// Init initializes the table node
func (n *TableNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	if maxRows, ok := config["maxRows"].(float64); ok {
		n.maxRows = int(maxRows)
	}
	if pagination, ok := config["pagination"].(bool); ok {
		n.pagination = pagination
	}
	if rowsPerPage, ok := config["rowsPerPage"].(float64); ok {
		n.rowsPerPage = int(rowsPerPage)
	}
	if searchable, ok := config["searchable"].(bool); ok {
		n.searchable = searchable
	}
	if exportable, ok := config["exportable"].(bool); ok {
		n.exportable = exportable
	}
	if striped, ok := config["striped"].(bool); ok {
		n.striped = striped
	}
	if bordered, ok := config["bordered"].(bool); ok {
		n.bordered = bordered
	}
	if compact, ok := config["compact"].(bool); ok {
		n.compact = compact
	}

	// Parse columns
	if columns, ok := config["columns"].([]interface{}); ok {
		for _, col := range columns {
			if colMap, ok := col.(map[string]interface{}); ok {
				column := TableColumn{
					Field:      colMap["field"].(string),
					Header:     colMap["header"].(string),
					Sortable:   true,
					Filterable: true,
				}
				if width, ok := colMap["width"].(string); ok {
					column.Width = width
				}
				if sortable, ok := colMap["sortable"].(bool); ok {
					column.Sortable = sortable
				}
				if filterable, ok := colMap["filterable"].(bool); ok {
					column.Filterable = filterable
				}
				if format, ok := colMap["format"].(string); ok {
					column.Format = format
				}
				n.columns = append(n.columns, column)
			}
		}
	}

	return nil
}

// Execute processes incoming messages and updates the table
func (n *TableNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract row data from payload
	if rows, ok := msg.Payload["rows"].([]interface{}); ok {
		// Replace all rows
		n.rows = make([]map[string]interface{}, 0)
		for _, row := range rows {
			if rowMap, ok := row.(map[string]interface{}); ok {
				n.rows = append(n.rows, rowMap)
			}
		}
	} else {
		// Add single row
		n.rows = append(n.rows, msg.Payload)
	}

	// Limit rows to maxRows
	if len(n.rows) > n.maxRows {
		n.rows = n.rows[len(n.rows)-n.maxRows:]
	}

	// Update dashboard
	if n.manager != nil {
		tableData := map[string]interface{}{
			"columns":     n.columns,
			"rows":        n.rows,
			"pagination":  n.pagination,
			"rowsPerPage": n.rowsPerPage,
			"searchable":  n.searchable,
			"exportable":  n.exportable,
			"striped":     n.striped,
			"bordered":    n.bordered,
			"compact":     n.compact,
		}
		n.manager.UpdateWidget(n.id, tableData)
	}

	return node.Message{}, nil
}

// SetManager sets the dashboard manager
func (n *TableNode) SetManager(manager *Manager) {
	n.manager = manager
}
