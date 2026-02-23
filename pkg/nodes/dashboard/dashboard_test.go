package dashboard

import (
	"context"
	"testing"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDashboardManager_RegisterWidget(t *testing.T) {
	manager := NewManager()

	widget := &Widget{
		ID:    "test-widget-1",
		Type:  WidgetTypeChart,
		Label: "Test Chart",
		Group: "test-group",
		Tab:   "test-tab",
	}

	err := manager.RegisterWidget(widget)
	require.NoError(t, err)

	retrieved, err := manager.GetWidget("test-widget-1")
	require.NoError(t, err)
	assert.Equal(t, widget.ID, retrieved.ID)
	assert.Equal(t, widget.Type, retrieved.Type)
}

func TestDashboardManager_UpdateWidget(t *testing.T) {
	manager := NewManager()

	widget := &Widget{
		ID:   "test-widget-2",
		Type: WidgetTypeGauge,
	}

	err := manager.RegisterWidget(widget)
	require.NoError(t, err)

	// Update widget value
	value := map[string]interface{}{"value": 75.5}
	err = manager.UpdateWidget("test-widget-2", value)
	require.NoError(t, err)

	// Verify update
	retrieved, err := manager.GetWidget("test-widget-2")
	require.NoError(t, err)
	assert.Equal(t, value, retrieved.LastValue)
	assert.False(t, retrieved.LastUpdated.IsZero())
}

func TestDashboardManager_Subscribe(t *testing.T) {
	manager := NewManager()

	widget := &Widget{
		ID:   "test-widget-3",
		Type: WidgetTypeText,
	}

	err := manager.RegisterWidget(widget)
	require.NoError(t, err)

	// Subscribe to updates
	ch, err := manager.Subscribe("test-widget-3")
	require.NoError(t, err)

	// Update widget
	go func() {
		time.Sleep(10 * time.Millisecond)
		manager.UpdateWidget("test-widget-3", "Hello Dashboard")
	}()

	// Wait for update
	select {
	case update := <-ch:
		assert.Equal(t, "test-widget-3", update.WidgetID)
		assert.Equal(t, "Hello Dashboard", update.Value)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive update")
	}
}

func TestChartNode_Execute(t *testing.T) {
	chart := NewChartNode()
	manager := NewManager()
	chart.SetManager(manager)

	err := chart.Init(map[string]interface{}{
		"id":          "chart-1",
		"label":       "Temperature Chart",
		"chartType":   "line",
		"maxDataSize": float64(50),
	})
	require.NoError(t, err)

	// Register widget with manager
	widget := &Widget{
		ID:   "chart-1",
		Type: WidgetTypeChart,
	}
	manager.RegisterWidget(widget)

	// Send data point
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"value": 22.5,
		},
	}

	_, err = chart.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify chart updated
	retrieved, err := manager.GetWidget("chart-1")
	require.NoError(t, err)
	assert.NotNil(t, retrieved.LastValue)
}

func TestGaugeNode_Execute(t *testing.T) {
	gauge := NewGaugeNode()
	manager := NewManager()
	gauge.SetManager(manager)

	err := gauge.Init(map[string]interface{}{
		"id":    "gauge-1",
		"label": "CPU Usage",
		"min":   float64(0),
		"max":   float64(100),
		"units": "%",
	})
	require.NoError(t, err)

	// Register widget
	widget := &Widget{
		ID:   "gauge-1",
		Type: WidgetTypeGauge,
	}
	manager.RegisterWidget(widget)

	// Send gauge value
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"value": 75.5,
		},
	}

	_, err = gauge.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify gauge updated
	retrieved, err := manager.GetWidget("gauge-1")
	require.NoError(t, err)
	assert.NotNil(t, retrieved.LastValue)

	gaugeData := retrieved.LastValue.(map[string]interface{})
	assert.Equal(t, 75.5, gaugeData["value"])
	assert.Equal(t, "%", gaugeData["units"])
}

func TestTextNode_Execute(t *testing.T) {
	text := NewTextNode()
	manager := NewManager()
	text.SetManager(manager)

	err := text.Init(map[string]interface{}{
		"id":    "text-1",
		"label": "Status",
	})
	require.NoError(t, err)

	// Register widget
	widget := &Widget{
		ID:   "text-1",
		Type: WidgetTypeText,
	}
	manager.RegisterWidget(widget)

	// Send text value
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"value": "System Online",
		},
	}

	_, err = text.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify text updated
	retrieved, err := manager.GetWidget("text-1")
	require.NoError(t, err)
	assert.NotNil(t, retrieved.LastValue)
}

func TestButtonNode_HandleClick(t *testing.T) {
	button := NewButtonNode()

	err := button.Init(map[string]interface{}{
		"id":          "button-1",
		"label":       "Trigger",
		"buttonLabel": "Click Me",
		"payload":     map[string]interface{}{"action": "start"},
	})
	require.NoError(t, err)

	err = button.Start(context.Background())
	require.NoError(t, err)

	// Simulate button click
	button.HandleClick()

	// Verify output message
	select {
	case msg := <-button.GetOutputChannel():
		assert.Equal(t, node.MessageTypeData, msg.Type)
		assert.Equal(t, "start", msg.Payload["action"])
		assert.NotNil(t, msg.Payload["timestamp"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive output message")
	}

	button.Cleanup()
}

func TestSliderNode_Execute(t *testing.T) {
	slider := NewSliderNode()

	err := slider.Init(map[string]interface{}{
		"id":    "slider-1",
		"label": "Volume",
		"min":   float64(0),
		"max":   float64(100),
		"step":  float64(5),
	})
	require.NoError(t, err)

	err = slider.Start(context.Background())
	require.NoError(t, err)

	// Simulate slider change
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"value": 75.0,
		},
	}

	_, err = slider.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify output message
	select {
	case outMsg := <-slider.GetOutputChannel():
		assert.Equal(t, 75.0, outMsg.Payload["value"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive output message")
	}

	slider.Cleanup()
}

func TestSwitchNode_Execute(t *testing.T) {
	switchNode := NewSwitchNode()

	err := switchNode.Init(map[string]interface{}{
		"id":       "switch-1",
		"label":    "Enable",
		"onValue":  "enabled",
		"offValue": "disabled",
	})
	require.NoError(t, err)

	err = switchNode.Start(context.Background())
	require.NoError(t, err)

	// Simulate switch to ON
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"value": true,
		},
	}

	_, err = switchNode.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify output message
	select {
	case outMsg := <-switchNode.GetOutputChannel():
		assert.Equal(t, "enabled", outMsg.Payload["value"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive output message")
	}

	switchNode.Cleanup()
}

func TestTextInputNode_Execute(t *testing.T) {
	textInput := NewTextInputNode()

	err := textInput.Init(map[string]interface{}{
		"id":          "input-1",
		"label":       "Name",
		"placeholder": "Enter your name",
	})
	require.NoError(t, err)

	err = textInput.Start(context.Background())
	require.NoError(t, err)

	// Simulate text input
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"value": "John Doe",
		},
	}

	_, err = textInput.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify output message
	select {
	case outMsg := <-textInput.GetOutputChannel():
		assert.Equal(t, "John Doe", outMsg.Payload["value"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive output message")
	}

	textInput.Cleanup()
}

func TestDropdownNode_Execute(t *testing.T) {
	dropdown := NewDropdownNode()

	options := []interface{}{
		map[string]interface{}{"label": "Option 1", "value": "opt1"},
		map[string]interface{}{"label": "Option 2", "value": "opt2"},
	}

	err := dropdown.Init(map[string]interface{}{
		"id":      "dropdown-1",
		"label":   "Select",
		"options": options,
	})
	require.NoError(t, err)

	err = dropdown.Start(context.Background())
	require.NoError(t, err)

	// Simulate selection
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"value": "opt1",
		},
	}

	_, err = dropdown.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify output message
	select {
	case outMsg := <-dropdown.GetOutputChannel():
		assert.Equal(t, "opt1", outMsg.Payload["value"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive output message")
	}

	dropdown.Cleanup()
}

func TestFormNode_Execute(t *testing.T) {
	form := NewFormNode()

	fields := []interface{}{
		map[string]interface{}{
			"name":     "username",
			"label":    "Username",
			"type":     "text",
			"required": true,
		},
		map[string]interface{}{
			"name":     "email",
			"label":    "Email",
			"type":     "email",
			"required": true,
		},
	}

	err := form.Init(map[string]interface{}{
		"id":     "form-1",
		"label":  "Registration",
		"fields": fields,
	})
	require.NoError(t, err)

	err = form.Start(context.Background())
	require.NoError(t, err)

	// Simulate form submission
	formData := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
	}

	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"formData": formData,
		},
	}

	_, err = form.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify output message
	select {
	case outMsg := <-form.GetOutputChannel():
		assert.Equal(t, "testuser", outMsg.Payload["username"])
		assert.Equal(t, "test@example.com", outMsg.Payload["email"])
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Did not receive output message")
	}

	form.Cleanup()
}

func TestTableNode_Execute(t *testing.T) {
	table := NewTableNode()
	manager := NewManager()
	table.SetManager(manager)

	columns := []interface{}{
		map[string]interface{}{"field": "name", "header": "Name"},
		map[string]interface{}{"field": "age", "header": "Age"},
	}

	err := table.Init(map[string]interface{}{
		"id":      "table-1",
		"label":   "Users",
		"columns": columns,
	})
	require.NoError(t, err)

	// Register widget
	widget := &Widget{
		ID:   "table-1",
		Type: WidgetTypeTable,
	}
	manager.RegisterWidget(widget)

	// Add table row
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"name": "John Doe",
			"age":  30,
		},
	}

	_, err = table.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify table updated
	retrieved, err := manager.GetWidget("table-1")
	require.NoError(t, err)
	assert.NotNil(t, retrieved.LastValue)
}

func TestNotificationNode_Execute(t *testing.T) {
	notification := NewNotificationNode()
	manager := NewManager()
	notification.SetManager(manager)

	err := notification.Init(map[string]interface{}{
		"id":       "notif-1",
		"label":    "Alert",
		"type":     "success",
		"duration": float64(5000),
	})
	require.NoError(t, err)

	// Register widget
	widget := &Widget{
		ID:   "notif-1",
		Type: WidgetTypeNotification,
	}
	manager.RegisterWidget(widget)

	// Send notification
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"title":   "Success",
			"message": "Operation completed successfully",
		},
	}

	_, err = notification.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify notification updated
	retrieved, err := manager.GetWidget("notif-1")
	require.NoError(t, err)
	assert.NotNil(t, retrieved.LastValue)

	notifData := retrieved.LastValue.(map[string]interface{})
	assert.Equal(t, "Success", notifData["title"])
	assert.Equal(t, "Operation completed successfully", notifData["message"])
}

func TestTemplateNode_Execute(t *testing.T) {
	template := NewTemplateNode()
	manager := NewManager()
	template.SetManager(manager)

	err := template.Init(map[string]interface{}{
		"id":       "template-1",
		"label":    "Custom",
		"template": "<div>Value: {{.value}}</div>",
	})
	require.NoError(t, err)

	// Register widget
	widget := &Widget{
		ID:   "template-1",
		Type: WidgetTypeTemplate,
	}
	manager.RegisterWidget(widget)

	// Send data
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"value": 42,
		},
	}

	_, err = template.Execute(context.Background(), msg)
	require.NoError(t, err)

	// Verify template rendered
	retrieved, err := manager.GetWidget("template-1")
	require.NoError(t, err)
	assert.NotNil(t, retrieved.LastValue)

	templateData := retrieved.LastValue.(map[string]interface{})
	assert.Contains(t, templateData["content"].(string), "Value: 42")
}
