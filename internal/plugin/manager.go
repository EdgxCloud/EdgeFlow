package plugin

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/edgeflow/edgeflow/internal/resources"
)

// Manager مدیریت پلاگین‌ها
type Manager struct {
	registry        *Registry
	nodeRegistry    *node.Registry
	resourceMonitor *resources.Monitor
	loader          *Loader

	loadOrder       []string          // ترتیب بارگذاری
	enabledPlugins  map[string]bool   // پلاگین‌های فعال

	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewManager ایجاد مدیر پلاگین
func NewManager(nodeRegistry *node.Registry, resourceMonitor *resources.Monitor) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	return &Manager{
		registry:        GetRegistry(),
		nodeRegistry:    nodeRegistry,
		resourceMonitor: resourceMonitor,
		loader:          NewLoader(),
		loadOrder:       make([]string, 0),
		enabledPlugins:  make(map[string]bool),
		ctx:             ctx,
		cancel:          cancel,
	}
}

// LoadPlugin بارگذاری پلاگین
func (m *Manager) LoadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// دریافت پلاگین از رجیستری
	plugin, err := m.registry.Get(name)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// بررسی اینکه قبلاً بارگذاری نشده باشد
	if plugin.IsLoaded() {
		return fmt.Errorf("plugin '%s' already loaded", name)
	}

	// بررسی منابع سیستم
	if m.resourceMonitor != nil {
		canLoad, reason := m.resourceMonitor.CanLoadModule(name, plugin.RequiredMemory())
		if !canLoad {
			return fmt.Errorf("cannot load plugin: %s", reason)
		}
	}

	// بررسی و بارگذاری وابستگی‌ها
	for _, dep := range plugin.Dependencies() {
		depPlugin, err := m.registry.Get(dep)
		if err != nil {
			return fmt.Errorf("dependency '%s' not found: %w", dep, err)
		}

		if !depPlugin.IsLoaded() {
			log.Printf("[PLUGIN] Loading dependency: %s", dep)
			if err := m.loadPluginInternal(depPlugin); err != nil {
				return fmt.Errorf("failed to load dependency '%s': %w", dep, err)
			}
		}
	}

	// بارگذاری پلاگین
	if err := m.loadPluginInternal(plugin); err != nil {
		return err
	}

	// افزودن به لیست فعال‌ها
	m.enabledPlugins[name] = true
	m.loadOrder = append(m.loadOrder, name)

	// ثبت در resource monitor
	if m.resourceMonitor != nil {
		m.resourceMonitor.EnableModule(name)
	}

	log.Printf("[PLUGIN] Loaded: %s v%s (%d nodes)", name, plugin.Version(), len(plugin.Nodes()))
	return nil
}

// loadPluginInternal بارگذاری داخلی پلاگین
func (m *Manager) loadPluginInternal(plugin Plugin) error {
	// تنظیم وضعیت
	if bp, ok := plugin.(*BasePlugin); ok {
		bp.SetStatus(StatusLoading)
	}

	// بارگذاری پلاگین
	if err := plugin.Load(); err != nil {
		if bp, ok := plugin.(*BasePlugin); ok {
			bp.SetError(err)
		}
		return fmt.Errorf("plugin load failed: %w", err)
	}

	// ثبت نودها در node registry
	for _, nodeDef := range plugin.Nodes() {
		nodeInfo := &node.NodeInfo{
			Type:        nodeDef.Type,
			Name:        nodeDef.Name,
			Category:    node.NodeType(nodeDef.Category),
			Description: nodeDef.Description,
			Icon:        nodeDef.Icon,
			Color:       nodeDef.Color,
			Factory:     nodeDef.Factory,
		}
		if err := m.nodeRegistry.Register(nodeInfo); err != nil {
			log.Printf("[WARN] Failed to register node '%s': %v", nodeDef.Type, err)
		}
	}

	// تنظیم وضعیت بارگذاری شده
	if bp, ok := plugin.(*BasePlugin); ok {
		bp.SetStatus(StatusLoaded)
	}

	return nil
}

// UnloadPlugin خارج کردن پلاگین از حافظه
func (m *Manager) UnloadPlugin(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// دریافت پلاگین
	plugin, err := m.registry.Get(name)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// بررسی بارگذاری
	if !plugin.IsLoaded() {
		return fmt.Errorf("plugin '%s' not loaded", name)
	}

	// بررسی وابستگی‌های معکوس
	for _, p := range m.registry.List() {
		if !p.IsLoaded() {
			continue
		}

		for _, dep := range p.Dependencies() {
			if dep == name {
				return fmt.Errorf("cannot unload: plugin '%s' depends on it", p.Name())
			}
		}
	}

	// خارج کردن از حافظه
	if err := m.unloadPluginInternal(plugin); err != nil {
		return err
	}

	// حذف از لیست فعال‌ها
	delete(m.enabledPlugins, name)

	// حذف از ترتیب بارگذاری
	for i, pname := range m.loadOrder {
		if pname == name {
			m.loadOrder = append(m.loadOrder[:i], m.loadOrder[i+1:]...)
			break
		}
	}

	// غیرفعال کردن در resource monitor
	if m.resourceMonitor != nil {
		m.resourceMonitor.DisableModule(name)
	}

	log.Printf("[PLUGIN] Unloaded: %s", name)
	return nil
}

// unloadPluginInternal خارج کردن داخلی از حافظه
func (m *Manager) unloadPluginInternal(plugin Plugin) error {
	// تنظیم وضعیت
	if bp, ok := plugin.(*BasePlugin); ok {
		bp.SetStatus(StatusUnloading)
	}

	// حذف نودها از registry
	for _, nodeDef := range plugin.Nodes() {
		// TODO: Unregister nodes (نیاز به تغییر در node.Registry)
		_ = nodeDef
	}

	// خارج کردن پلاگین
	if err := plugin.Unload(); err != nil {
		if bp, ok := plugin.(*BasePlugin); ok {
			bp.SetError(err)
		}
		return fmt.Errorf("plugin unload failed: %w", err)
	}

	// تنظیم وضعیت
	if bp, ok := plugin.(*BasePlugin); ok {
		bp.SetStatus(StatusNotLoaded)
	}

	return nil
}

// EnablePlugin فعال‌سازی پلاگین
func (m *Manager) EnablePlugin(name string) error {
	m.mu.Lock()
	m.enabledPlugins[name] = true
	m.mu.Unlock()

	return m.LoadPlugin(name)
}

// DisablePlugin غیرفعال‌سازی پلاگین
func (m *Manager) DisablePlugin(name string) error {
	m.mu.Lock()
	m.enabledPlugins[name] = false
	m.mu.Unlock()

	return m.UnloadPlugin(name)
}

// IsEnabled بررسی فعال بودن پلاگین
func (m *Manager) IsEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.enabledPlugins[name]
}

// LoadAll بارگذاری تمام پلاگین‌های فعال
func (m *Manager) LoadAll() error {
	m.mu.RLock()
	enabled := make([]string, 0, len(m.enabledPlugins))
	for name, isEnabled := range m.enabledPlugins {
		if isEnabled {
			enabled = append(enabled, name)
		}
	}
	m.mu.RUnlock()

	// بارگذاری به ترتیب
	for _, name := range enabled {
		if err := m.LoadPlugin(name); err != nil {
			log.Printf("[ERROR] Failed to load plugin '%s': %v", name, err)
			// ادامه با سایر پلاگین‌ها
		}
	}

	return nil
}

// UnloadAll خارج کردن تمام پلاگین‌ها از حافظه
func (m *Manager) UnloadAll() error {
	m.mu.RLock()
	loadOrder := make([]string, len(m.loadOrder))
	copy(loadOrder, m.loadOrder)
	m.mu.RUnlock()

	// خارج کردن به ترتیب معکوس
	for i := len(loadOrder) - 1; i >= 0; i-- {
		name := loadOrder[i]
		if err := m.UnloadPlugin(name); err != nil {
			log.Printf("[ERROR] Failed to unload plugin '%s': %v", name, err)
		}
	}

	return nil
}

// ReloadPlugin بارگذاری مجدد پلاگین
func (m *Manager) ReloadPlugin(name string) error {
	// خارج کردن
	if err := m.UnloadPlugin(name); err != nil {
		return err
	}

	// بارگذاری مجدد
	return m.LoadPlugin(name)
}

// CheckCompatibility بررسی سازگاری پلاگین با سیستم
func (m *Manager) CheckCompatibility(name string) (bool, string) {
	plugin, err := m.registry.Get(name)
	if err != nil {
		return false, "plugin not found"
	}

	// بررسی منابع
	if m.resourceMonitor != nil {
		canLoad, reason := m.resourceMonitor.CanLoadModule(name, plugin.RequiredMemory())
		if !canLoad {
			return false, reason
		}
	}

	// بررسی وابستگی‌ها
	for _, dep := range plugin.Dependencies() {
		_, err := m.registry.Get(dep)
		if err != nil {
			return false, fmt.Sprintf("missing dependency: %s", dep)
		}
	}

	return true, "compatible"
}

// GetPluginInfo دریافت اطلاعات پلاگین
func (m *Manager) GetPluginInfo(name string) (PluginInfo, error) {
	plugin, err := m.registry.Get(name)
	if err != nil {
		return PluginInfo{}, err
	}

	compatible, reason := m.CheckCompatibility(name)
	return ToPluginInfo(plugin, compatible, reason), nil
}

// ListPlugins لیست تمام پلاگین‌ها با اطلاعات
func (m *Manager) ListPlugins() []PluginInfo {
	plugins := m.registry.List()
	infos := make([]PluginInfo, 0, len(plugins))

	for _, plugin := range plugins {
		compatible, reason := m.CheckCompatibility(plugin.Name())
		infos = append(infos, ToPluginInfo(plugin, compatible, reason))
	}

	return infos
}

// GetStats آمار پلاگین‌ها
func (m *Manager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	totalNodes := 0
	for _, plugin := range m.registry.ListLoaded() {
		totalNodes += len(plugin.Nodes())
	}

	return map[string]interface{}{
		"total_plugins":      m.registry.Count(),
		"loaded_plugins":     m.registry.CountLoaded(),
		"enabled_plugins":    len(m.enabledPlugins),
		"total_nodes":        totalNodes,
		"load_order":         m.loadOrder,
		"memory_available_mb": m.getAvailableMemoryMB(),
	}
}

// getAvailableMemoryMB دریافت حافظه موجود
func (m *Manager) getAvailableMemoryMB() uint64 {
	if m.resourceMonitor == nil {
		return 0
	}

	stats := m.resourceMonitor.GetStats()
	return stats.MemoryAvailable / 1024 / 1024
}

// Shutdown خاموش کردن مدیر
func (m *Manager) Shutdown() error {
	m.cancel()
	return m.UnloadAll()
}

// SetResourceMonitor تنظیم resource monitor
func (m *Manager) SetResourceMonitor(monitor *resources.Monitor) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.resourceMonitor = monitor
}

// GetLoadOrder دریافت ترتیب بارگذاری
func (m *Manager) GetLoadOrder() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	order := make([]string, len(m.loadOrder))
	copy(order, m.loadOrder)
	return order
}
