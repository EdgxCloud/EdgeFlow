package dashboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// WidgetType defines the type of dashboard widget
type WidgetType string

const (
	WidgetTypeChart        WidgetType = "chart"
	WidgetTypeGauge        WidgetType = "gauge"
	WidgetTypeText         WidgetType = "text"
	WidgetTypeButton       WidgetType = "button"
	WidgetTypeSlider       WidgetType = "slider"
	WidgetTypeSwitch       WidgetType = "switch"
	WidgetTypeTextInput    WidgetType = "text-input"
	WidgetTypeDropdown     WidgetType = "dropdown"
	WidgetTypeForm         WidgetType = "form"
	WidgetTypeDatePicker   WidgetType = "date-picker"
	WidgetTypeNotification WidgetType = "notification"
	WidgetTypeTemplate     WidgetType = "template"
	WidgetTypeTable        WidgetType = "table"
	WidgetTypeAudio        WidgetType = "audio"
)

// Widget represents a dashboard widget configuration
type Widget struct {
	ID          string                 `json:"id"`
	Type        WidgetType             `json:"type"`
	Label       string                 `json:"label"`
	Group       string                 `json:"group"`
	Tab         string                 `json:"tab"`
	Order       int                    `json:"order"`
	Width       int                    `json:"width"`
	Height      int                    `json:"height"`
	Config      map[string]interface{} `json:"config"`
	LastValue   interface{}            `json:"lastValue,omitempty"`
	LastUpdated time.Time              `json:"lastUpdated,omitempty"`
}

// DashboardGroup represents a group of widgets
type DashboardGroup struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Tab   string `json:"tab"`
	Order int    `json:"order"`
	Width int    `json:"width"`
}

// DashboardTab represents a tab in the dashboard
type DashboardTab struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Order int    `json:"order"`
}

// DashboardUpdate represents a real-time widget update
type DashboardUpdate struct {
	WidgetID  string      `json:"widgetId"`
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
}

// DashboardEvent represents a user interaction event
type DashboardEvent struct {
	WidgetID  string                 `json:"widgetId"`
	EventType string                 `json:"eventType"`
	Value     interface{}            `json:"value"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// Manager manages dashboard widgets and their state
type Manager struct {
	widgets      map[string]*Widget
	groups       map[string]*DashboardGroup
	tabs         map[string]*DashboardTab
	subscribers  map[string][]chan DashboardUpdate
	eventHandler func(DashboardEvent) error
	mu           sync.RWMutex
}

// NewManager creates a new dashboard manager
func NewManager() *Manager {
	return &Manager{
		widgets:     make(map[string]*Widget),
		groups:      make(map[string]*DashboardGroup),
		tabs:        make(map[string]*DashboardTab),
		subscribers: make(map[string][]chan DashboardUpdate),
	}
}

// RegisterWidget registers a new widget
func (m *Manager) RegisterWidget(widget *Widget) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if widget.ID == "" {
		return fmt.Errorf("widget ID is required")
	}

	m.widgets[widget.ID] = widget
	return nil
}

// UnregisterWidget removes a widget
func (m *Manager) UnregisterWidget(widgetID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.widgets, widgetID)
	delete(m.subscribers, widgetID)
	return nil
}

// UpdateWidget updates a widget's value
func (m *Manager) UpdateWidget(widgetID string, value interface{}) error {
	m.mu.Lock()
	widget, exists := m.widgets[widgetID]
	if !exists {
		m.mu.Unlock()
		return fmt.Errorf("widget %s not found", widgetID)
	}

	widget.LastValue = value
	widget.LastUpdated = time.Now()

	// Get subscribers for this widget
	subscribers := m.subscribers[widgetID]
	m.mu.Unlock()

	// Send update to all subscribers
	update := DashboardUpdate{
		WidgetID:  widgetID,
		Value:     value,
		Timestamp: time.Now(),
	}

	for _, ch := range subscribers {
		select {
		case ch <- update:
		default:
			// Skip if channel is full
		}
	}

	return nil
}

// GetWidget retrieves a widget by ID
func (m *Manager) GetWidget(widgetID string) (*Widget, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	widget, exists := m.widgets[widgetID]
	if !exists {
		return nil, fmt.Errorf("widget %s not found", widgetID)
	}

	return widget, nil
}

// GetAllWidgets returns all registered widgets
func (m *Manager) GetAllWidgets() []*Widget {
	m.mu.RLock()
	defer m.mu.RUnlock()

	widgets := make([]*Widget, 0, len(m.widgets))
	for _, widget := range m.widgets {
		widgets = append(widgets, widget)
	}

	return widgets
}

// Subscribe creates a subscription channel for widget updates
func (m *Manager) Subscribe(widgetID string) (<-chan DashboardUpdate, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.widgets[widgetID]; !exists {
		return nil, fmt.Errorf("widget %s not found", widgetID)
	}

	ch := make(chan DashboardUpdate, 100)
	m.subscribers[widgetID] = append(m.subscribers[widgetID], ch)

	return ch, nil
}

// RegisterGroup registers a dashboard group
func (m *Manager) RegisterGroup(group *DashboardGroup) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if group.ID == "" {
		return fmt.Errorf("group ID is required")
	}

	m.groups[group.ID] = group
	return nil
}

// RegisterTab registers a dashboard tab
func (m *Manager) RegisterTab(tab *DashboardTab) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if tab.ID == "" {
		return fmt.Errorf("tab ID is required")
	}

	m.tabs[tab.ID] = tab
	return nil
}

// GetGroups returns all dashboard groups
func (m *Manager) GetGroups() []*DashboardGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()

	groups := make([]*DashboardGroup, 0, len(m.groups))
	for _, group := range m.groups {
		groups = append(groups, group)
	}

	return groups
}

// GetTabs returns all dashboard tabs
func (m *Manager) GetTabs() []*DashboardTab {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tabs := make([]*DashboardTab, 0, len(m.tabs))
	for _, tab := range m.tabs {
		tabs = append(tabs, tab)
	}

	return tabs
}

// SetEventHandler sets the event handler for dashboard interactions
func (m *Manager) SetEventHandler(handler func(DashboardEvent) error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.eventHandler = handler
}

// HandleEvent processes a dashboard event
func (m *Manager) HandleEvent(event DashboardEvent) error {
	m.mu.RLock()
	handler := m.eventHandler
	m.mu.RUnlock()

	if handler != nil {
		return handler(event)
	}

	return nil
}

// BaseWidget provides common functionality for all dashboard widgets
type BaseWidget struct {
	id          string
	widgetType  WidgetType
	label       string
	group       string
	tab         string
	config      map[string]interface{}
	manager     *Manager
	outputChan  chan node.Message
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewBaseWidget creates a new base widget
func NewBaseWidget(widgetType WidgetType) *BaseWidget {
	return &BaseWidget{
		widgetType: widgetType,
		config:     make(map[string]interface{}),
		outputChan: make(chan node.Message, 100),
	}
}

// Init initializes the base widget
func (w *BaseWidget) Init(config map[string]interface{}) error {
	w.config = config

	// Parse common configuration
	if id, ok := config["id"].(string); ok {
		w.id = id
	}
	if label, ok := config["label"].(string); ok {
		w.label = label
	}
	if group, ok := config["group"].(string); ok {
		w.group = group
	}
	if tab, ok := config["tab"].(string); ok {
		w.tab = tab
	}

	return nil
}

// Start starts the widget
func (w *BaseWidget) Start(ctx context.Context) error {
	w.ctx, w.cancel = context.WithCancel(ctx)
	return nil
}

// Cleanup cleans up widget resources
func (w *BaseWidget) Cleanup() error {
	if w.cancel != nil {
		w.cancel()
	}
	close(w.outputChan)
	return nil
}

// GetOutputChannel returns the output channel
func (w *BaseWidget) GetOutputChannel() <-chan node.Message {
	return w.outputChan
}

// SendOutput sends a message to the output channel
func (w *BaseWidget) SendOutput(msg node.Message) {
	select {
	case w.outputChan <- msg:
	case <-w.ctx.Done():
	default:
		// Skip if channel full
	}
}

// GetID returns the widget ID
func (w *BaseWidget) GetID() string {
	return w.id
}

// GetType returns the widget type
func (w *BaseWidget) GetType() WidgetType {
	return w.widgetType
}

// GetConfig returns the widget configuration
func (w *BaseWidget) GetConfig() map[string]interface{} {
	return w.config
}
