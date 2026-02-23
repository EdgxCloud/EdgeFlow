package dashboard

import (
	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// RegisterAll registers all dashboard widgets with the node registry
func RegisterAll(registry *node.Registry) error {
	// Register Chart Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-chart",
		Name:        "Dashboard Chart",
		Category:    node.NodeTypeOutput,
		Description: "Display data as charts (line, bar, pie, histogram, scatter)",
		Icon:        "chart-line",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "chartType", Label: "Chart Type", Type: "select", Default: "line",
				Options: []string{"line", "bar", "pie", "histogram", "scatter"}},
			{Name: "xAxisLabel", Label: "X-Axis Label", Type: "string"},
			{Name: "yAxisLabel", Label: "Y-Axis Label", Type: "string"},
			{Name: "maxDataSize", Label: "Max Data Points", Type: "number", Default: float64(100)},
			{Name: "legend", Label: "Show Legend", Type: "boolean", Default: true},
		},
		Factory: func() node.Executor {
			return NewChartNode()
		},
	}); err != nil {
		return err
	}

	// Register Gauge Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-gauge",
		Name:        "Dashboard Gauge",
		Category:    node.NodeTypeOutput,
		Description: "Display numeric values as gauges with colored sectors",
		Icon:        "gauge",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "min", Label: "Minimum", Type: "number", Default: float64(0)},
			{Name: "max", Label: "Maximum", Type: "number", Default: float64(100)},
			{Name: "units", Label: "Units", Type: "string"},
			{Name: "showValue", Label: "Show Value", Type: "boolean", Default: true},
			{Name: "showMinMax", Label: "Show Min/Max", Type: "boolean", Default: true},
		},
		Factory: func() node.Executor {
			return NewGaugeNode()
		},
	}); err != nil {
		return err
	}

	// Register Text Display Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-text",
		Name:        "Dashboard Text",
		Category:    node.NodeTypeOutput,
		Description: "Display text on the dashboard",
		Icon:        "text",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "format", Label: "Format", Type: "string", Default: "{{msg.payload}}"},
			{Name: "fontSize", Label: "Font Size", Type: "number", Default: float64(14)},
			{Name: "fontColor", Label: "Font Color", Type: "string", Default: "#000000"},
			{Name: "layout", Label: "Layout", Type: "select", Default: "row-spread",
				Options: []string{"row-spread", "row-left", "row-center", "row-right", "col-center"}},
		},
		Factory: func() node.Executor {
			return NewTextNode()
		},
	}); err != nil {
		return err
	}

	// Register Button Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-button",
		Name:        "Dashboard Button",
		Category:    node.NodeTypeInput,
		Description: "Interactive button on the dashboard",
		Icon:        "hand-pointer",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "buttonLabel", Label: "Button Label", Type: "string", Default: "Button"},
			{Name: "payload", Label: "Payload", Type: "payload"},
			{Name: "topic", Label: "Topic", Type: "string"},
			{Name: "icon", Label: "Icon", Type: "string"},
			{Name: "bgColor", Label: "Background Color", Type: "string", Default: "#3b82f6"},
			{Name: "fgColor", Label: "Text Color", Type: "string", Default: "#ffffff"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewButtonNode()
		},
	}); err != nil {
		return err
	}

	// Register Slider Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-slider",
		Name:        "Dashboard Slider",
		Category:    node.NodeTypeInput,
		Description: "Interactive slider input",
		Icon:        "sliders",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "min", Label: "Minimum", Type: "number", Default: float64(0)},
			{Name: "max", Label: "Maximum", Type: "number", Default: float64(100)},
			{Name: "step", Label: "Step", Type: "number", Default: float64(1)},
			{Name: "value", Label: "Initial Value", Type: "number", Default: float64(50)},
			{Name: "units", Label: "Units", Type: "string"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "number"},
		},
		Factory: func() node.Executor {
			return NewSliderNode()
		},
	}); err != nil {
		return err
	}

	// Register Switch Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-switch",
		Name:        "Dashboard Switch",
		Category:    node.NodeTypeInput,
		Description: "Interactive switch/toggle input",
		Icon:        "toggle-on",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "onValue", Label: "On Value", Type: "any", Default: true},
			{Name: "offValue", Label: "Off Value", Type: "any", Default: false},
			{Name: "state", Label: "Initial State", Type: "boolean", Default: false},
			{Name: "onLabel", Label: "On Label", Type: "string", Default: "On"},
			{Name: "offLabel", Label: "Off Label", Type: "string", Default: "Off"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewSwitchNode()
		},
	}); err != nil {
		return err
	}

	// Register Text Input Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-text-input",
		Name:        "Dashboard Text Input",
		Category:    node.NodeTypeInput,
		Description: "Text input field",
		Icon:        "keyboard",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "placeholder", Label: "Placeholder", Type: "string", Default: "Enter text..."},
			{Name: "mode", Label: "Input Mode", Type: "select", Default: "text",
				Options: []string{"text", "password", "email", "number"}},
			{Name: "delay", Label: "Delay (ms)", Type: "number", Default: float64(300)},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "string"},
		},
		Factory: func() node.Executor {
			return NewTextInputNode()
		},
	}); err != nil {
		return err
	}

	// Register Dropdown Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-dropdown",
		Name:        "Dashboard Dropdown",
		Category:    node.NodeTypeInput,
		Description: "Dropdown select input",
		Icon:        "list",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "options", Label: "Options", Type: "array", Required: true},
			{Name: "placeholder", Label: "Placeholder", Type: "string", Default: "Select an option..."},
			{Name: "multiple", Label: "Multiple Selection", Type: "boolean", Default: false},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewDropdownNode()
		},
	}); err != nil {
		return err
	}

	// Register Form Builder Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-form",
		Name:        "Dashboard Form",
		Category:    node.NodeTypeInput,
		Description: "Form builder with multiple input fields",
		Icon:        "file-alt",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "fields", Label: "Form Fields", Type: "array", Required: true},
			{Name: "submitLabel", Label: "Submit Button Label", Type: "string", Default: "Submit"},
			{Name: "resetLabel", Label: "Reset Button Label", Type: "string", Default: "Reset"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object"},
		},
		Factory: func() node.Executor {
			return NewFormNode()
		},
	}); err != nil {
		return err
	}

	// Register Date Picker Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-date-picker",
		Name:        "Dashboard Date Picker",
		Category:    node.NodeTypeInput,
		Description: "Date and time picker input",
		Icon:        "calendar",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "mode", Label: "Mode", Type: "select", Default: "datetime",
				Options: []string{"date", "time", "datetime"}},
			{Name: "format", Label: "Date Format", Type: "string", Default: "2006-01-02 15:04:05"},
			{Name: "enableTime", Label: "Enable Time", Type: "boolean", Default: true},
			{Name: "enableSeconds", Label: "Enable Seconds", Type: "boolean", Default: false},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object"},
		},
		Factory: func() node.Executor {
			return NewDatePickerNode()
		},
	}); err != nil {
		return err
	}

	// Register Notification Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-notification",
		Name:        "Dashboard Notification",
		Category:    node.NodeTypeOutput,
		Description: "Display toast notifications on the dashboard",
		Icon:        "bell",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "type", Label: "Type", Type: "select", Default: "info",
				Options: []string{"success", "info", "warning", "error"}},
			{Name: "duration", Label: "Duration (ms)", Type: "number", Default: float64(3000)},
			{Name: "position", Label: "Position", Type: "select", Default: "top-right",
				Options: []string{"top-left", "top-center", "top-right", "bottom-left", "bottom-center", "bottom-right"}},
			{Name: "closable", Label: "Closable", Type: "boolean", Default: true},
		},
		Factory: func() node.Executor {
			return NewNotificationNode()
		},
	}); err != nil {
		return err
	}

	// Register Template Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-template",
		Name:        "Dashboard Template",
		Category:    node.NodeTypeOutput,
		Description: "Display custom HTML/template content",
		Icon:        "code",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "template", Label: "Template", Type: "string", Default: "<p>{{.value}}</p>"},
			{Name: "allowHTML", Label: "Allow Raw HTML", Type: "boolean", Default: false},
		},
		Factory: func() node.Executor {
			return NewTemplateNode()
		},
	}); err != nil {
		return err
	}

	// Register Color Picker Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-color-picker",
		Name:        "Dashboard Color Picker",
		Category:    node.NodeTypeInput,
		Description: "Color picker input widget for dashboard",
		Icon:        "palette",
		Color:       "#E91E63",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "format", Label: "Format", Type: "select", Default: "hex",
				Options: []string{"hex", "rgb", "hsl"}},
			{Name: "defaultColor", Label: "Default Color", Type: "string", Default: "#3b82f6"},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Set color programmatically"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Color value in all formats"},
		},
		Factory: func() node.Executor {
			return NewColorPickerExecutor()
		},
	}); err != nil {
		return err
	}

	// Register Table Widget
	if err := registry.Register(&node.NodeInfo{
		Type:        "dashboard-table",
		Name:        "Dashboard Table",
		Category:    node.NodeTypeOutput,
		Description: "Display data in a table with sorting and filtering",
		Icon:        "table",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{Name: "id", Label: "Widget ID", Type: "string", Required: true},
			{Name: "label", Label: "Label", Type: "string", Required: true},
			{Name: "group", Label: "Group", Type: "string"},
			{Name: "tab", Label: "Tab", Type: "string"},
			{Name: "columns", Label: "Columns", Type: "array", Required: true},
			{Name: "maxRows", Label: "Max Rows", Type: "number", Default: float64(1000)},
			{Name: "pagination", Label: "Enable Pagination", Type: "boolean", Default: true},
			{Name: "rowsPerPage", Label: "Rows Per Page", Type: "number", Default: float64(10)},
			{Name: "searchable", Label: "Searchable", Type: "boolean", Default: true},
			{Name: "exportable", Label: "Exportable", Type: "boolean", Default: true},
			{Name: "striped", Label: "Striped Rows", Type: "boolean", Default: true},
			{Name: "bordered", Label: "Bordered", Type: "boolean", Default: false},
			{Name: "compact", Label: "Compact Mode", Type: "boolean", Default: false},
		},
		Factory: func() node.Executor {
			return NewTableNode()
		},
	}); err != nil {
		return err
	}

	return nil
}
