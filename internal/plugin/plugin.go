package plugin

import (
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// Plugin رابط اصلی برای تمام پلاگین‌ها
type Plugin interface {
	// اطلاعات پلاگین
	Name() string
	Version() string
	Description() string
	Author() string
	Category() Category

	// مدیریت چرخه حیات
	Load() error
	Unload() error
	IsLoaded() bool

	// نودهای ارائه شده
	Nodes() []NodeDefinition

	// نیازمندی‌های سیستم
	RequiredMemory() uint64  // بایت
	RequiredDisk() uint64    // بایت
	Dependencies() []string  // نام پلاگین‌های وابسته

	// تنظیمات
	DefaultConfig() map[string]interface{}
	ValidateConfig(config map[string]interface{}) error
}

// Category دسته‌بندی پلاگین
type Category string

const (
	CategoryCore       Category = "core"
	CategoryNetwork    Category = "network"
	CategoryGPIO       Category = "gpio"
	CategoryDatabase   Category = "database"
	CategoryMessaging  Category = "messaging"
	CategoryAI         Category = "ai"
	CategoryIndustrial Category = "industrial"
	CategoryAdvanced   Category = "advanced"
	CategoryUI         Category = "ui"
)

// NodeDefinition تعریف نود ارائه شده توسط پلاگین
type NodeDefinition struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Category    string                 `json:"category"`
	Description string                 `json:"description"`
	Icon        string                 `json:"icon"`
	Color       string                 `json:"color"`
	Inputs      int                    `json:"inputs"`
	Outputs     int                    `json:"outputs"`
	Factory     node.NodeFactory       `json:"-"`
	Config      map[string]interface{} `json:"config"`
}

// Metadata اطلاعات متا پلاگین
type Metadata struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	Category     Category               `json:"category"`
	License      string                 `json:"license"`
	Homepage     string                 `json:"homepage"`
	Repository   string                 `json:"repository"`
	Keywords     []string               `json:"keywords"`
	MinEdgeFlow  string                 `json:"min_edgeflow"`
	Config       map[string]interface{} `json:"config"`
	Dependencies []string               `json:"dependencies"`
}

// Status وضعیت پلاگین
type Status string

const (
	StatusNotLoaded Status = "not_loaded"
	StatusLoading   Status = "loading"
	StatusLoaded    Status = "loaded"
	StatusUnloading Status = "unloading"
	StatusError     Status = "error"
)

// BasePlugin پیاده‌سازی پایه برای پلاگین‌ها
type BasePlugin struct {
	metadata     Metadata
	status       Status
	loadedAt     time.Time
	unloadedAt   time.Time
	errorMessage string
	config       map[string]interface{}
	mu           sync.RWMutex
}

// NewBasePlugin ایجاد پلاگین پایه
func NewBasePlugin(metadata Metadata) *BasePlugin {
	return &BasePlugin{
		metadata: metadata,
		status:   StatusNotLoaded,
		config:   make(map[string]interface{}),
	}
}

// Name نام پلاگین
func (p *BasePlugin) Name() string {
	return p.metadata.Name
}

// Version نسخه پلاگین
func (p *BasePlugin) Version() string {
	return p.metadata.Version
}

// Description توضیحات پلاگین
func (p *BasePlugin) Description() string {
	return p.metadata.Description
}

// Author نویسنده پلاگین
func (p *BasePlugin) Author() string {
	return p.metadata.Author
}

// Category دسته‌بندی پلاگین
func (p *BasePlugin) Category() Category {
	return p.metadata.Category
}

// GetMetadata دریافت تمام اطلاعات متا
func (p *BasePlugin) GetMetadata() Metadata {
	return p.metadata
}

// SetStatus تنظیم وضعیت
func (p *BasePlugin) SetStatus(status Status) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = status

	if status == StatusLoaded {
		p.loadedAt = time.Now()
	} else if status == StatusNotLoaded {
		p.unloadedAt = time.Now()
	}
}

// GetStatus دریافت وضعیت
func (p *BasePlugin) GetStatus() Status {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.status
}

// IsLoaded بررسی بارگذاری شدن
func (p *BasePlugin) IsLoaded() bool {
	return p.GetStatus() == StatusLoaded
}

// SetError تنظیم خطا
func (p *BasePlugin) SetError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status = StatusError
	if err != nil {
		p.errorMessage = err.Error()
	}
}

// GetError دریافت خطا
func (p *BasePlugin) GetError() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.errorMessage
}

// SetConfig تنظیم پیکربندی
func (p *BasePlugin) SetConfig(config map[string]interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.config = config
}

// GetConfig دریافت پیکربندی
func (p *BasePlugin) GetConfig() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// LoadedAt زمان بارگذاری
func (p *BasePlugin) LoadedAt() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.loadedAt
}

// UnloadedAt زمان خارج شدن از حافظه
func (p *BasePlugin) UnloadedAt() time.Time {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.unloadedAt
}

// DefaultConfig پیکربندی پیش‌فرض
func (p *BasePlugin) DefaultConfig() map[string]interface{} {
	return p.metadata.Config
}

// ValidateConfig اعتبارسنجی پیکربندی
func (p *BasePlugin) ValidateConfig(config map[string]interface{}) error {
	// پیاده‌سازی پایه - پلاگین‌ها می‌توانند override کنند
	return nil
}

// Dependencies لیست پلاگین‌های وابسته
func (p *BasePlugin) Dependencies() []string {
	return p.metadata.Dependencies
}

// Load بارگذاری پلاگین (پیاده‌سازی پیش‌فرض - پلاگین‌ها می‌توانند override کنند)
func (p *BasePlugin) Load() error {
	return nil
}

// Unload خارج کردن پلاگین از حافظه (پیاده‌سازی پیش‌فرض - پلاگین‌ها می‌توانند override کنند)
func (p *BasePlugin) Unload() error {
	return nil
}

// Nodes لیست نودهای ارائه شده (پیاده‌سازی پیش‌فرض - پلاگین‌ها باید override کنند)
func (p *BasePlugin) Nodes() []NodeDefinition {
	return []NodeDefinition{}
}

// RequiredMemory حافظه مورد نیاز (پیاده‌سازی پیش‌فرض)
func (p *BasePlugin) RequiredMemory() uint64 {
	return 0
}

// RequiredDisk فضای دیسک مورد نیاز (پیاده‌سازی پیش‌فرض)
func (p *BasePlugin) RequiredDisk() uint64 {
	return 0
}

// PluginInfo اطلاعات کامل پلاگین برای API
type PluginInfo struct {
	Name             string                 `json:"name"`
	Version          string                 `json:"version"`
	Description      string                 `json:"description"`
	Author           string                 `json:"author"`
	Category         Category               `json:"category"`
	Status           Status                 `json:"status"`
	LoadedAt         *time.Time             `json:"loaded_at,omitempty"`
	UnloadedAt       *time.Time             `json:"unloaded_at,omitempty"`
	Error            string                 `json:"error,omitempty"`
	RequiredMemoryMB uint64                 `json:"required_memory_mb"`
	RequiredDiskMB   uint64                 `json:"required_disk_mb"`
	Dependencies     []string               `json:"dependencies"`
	Nodes            []NodeDefinition       `json:"nodes"`
	Config           map[string]interface{} `json:"config"`
	Compatible       bool                   `json:"compatible"`
	CompatibleReason string                 `json:"compatible_reason,omitempty"`
}

// ToPluginInfo تبدیل به PluginInfo
func ToPluginInfo(p Plugin, compatible bool, reason string) PluginInfo {
	info := PluginInfo{
		Name:             p.Name(),
		Version:          p.Version(),
		Description:      p.Description(),
		Author:           p.Author(),
		Category:         p.Category(),
		RequiredMemoryMB: p.RequiredMemory() / 1024 / 1024,
		RequiredDiskMB:   p.RequiredDisk() / 1024 / 1024,
		Dependencies:     p.Dependencies(),
		Nodes:            p.Nodes(),
		Config:           p.DefaultConfig(),
		Compatible:       compatible,
		CompatibleReason: reason,
	}

	// اطلاعات وضعیت اگر BasePlugin باشد
	if bp, ok := p.(*BasePlugin); ok {
		info.Status = bp.GetStatus()
		info.Error = bp.GetError()

		if !bp.LoadedAt().IsZero() {
			t := bp.LoadedAt()
			info.LoadedAt = &t
		}
		if !bp.UnloadedAt().IsZero() {
			t := bp.UnloadedAt()
			info.UnloadedAt = &t
		}
	} else {
		if p.IsLoaded() {
			info.Status = StatusLoaded
		} else {
			info.Status = StatusNotLoaded
		}
	}

	return info
}

// Registry رجیستری جهانی پلاگین‌ها
type Registry struct {
	plugins map[string]Plugin
	mu      sync.RWMutex
}

// globalRegistry رجیستری جهانی
var globalRegistry = &Registry{
	plugins: make(map[string]Plugin),
}

// GetRegistry دریافت رجیستری جهانی
func GetRegistry() *Registry {
	return globalRegistry
}

// Register ثبت پلاگین در رجیستری
func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := plugin.Name()
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' already registered", name)
	}

	r.plugins[name] = plugin
	return nil
}

// Unregister حذف پلاگین از رجیستری
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	delete(r.plugins, name)
	return nil
}

// Get دریافت پلاگین
func (r *Registry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	return plugin, nil
}

// List لیست تمام پلاگین‌ها
func (r *Registry) List() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0, len(r.plugins))
	for _, p := range r.plugins {
		plugins = append(plugins, p)
	}

	return plugins
}

// ListByCategory لیست پلاگین‌ها بر اساس دسته‌بندی
func (r *Registry) ListByCategory(category Category) []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0)
	for _, p := range r.plugins {
		if p.Category() == category {
			plugins = append(plugins, p)
		}
	}

	return plugins
}

// ListLoaded لیست پلاگین‌های بارگذاری شده
func (r *Registry) ListLoaded() []Plugin {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugins := make([]Plugin, 0)
	for _, p := range r.plugins {
		if p.IsLoaded() {
			plugins = append(plugins, p)
		}
	}

	return plugins
}

// Count تعداد پلاگین‌ها
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.plugins)
}

// CountLoaded تعداد پلاگین‌های بارگذاری شده
func (r *Registry) CountLoaded() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, p := range r.plugins {
		if p.IsLoaded() {
			count++
		}
	}
	return count
}
