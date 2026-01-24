package plugin

import (
	"github.com/edgeflow/edgeflow/internal/node"
	coreNodes "github.com/edgeflow/edgeflow/pkg/nodes/core"
)

// CorePlugin پلاگین هسته (همیشه بارگذاری می‌شود)
type CorePlugin struct {
	*BasePlugin
}

// NewCorePlugin ایجاد پلاگین هسته
func NewCorePlugin() *CorePlugin {
	metadata := Metadata{
		Name:        "core",
		Version:     "1.0.0",
		Description: "Core nodes for EdgeFlow - always required",
		Author:      "EdgeFlow Team",
		Category:    CategoryCore,
		License:     "Apache-2.0",
		Config:      make(map[string]interface{}),
	}

	return &CorePlugin{
		BasePlugin: NewBasePlugin(metadata),
	}
}

// Load بارگذاری پلاگین
func (p *CorePlugin) Load() error {
	// هیچ منطق خاصی نیاز نیست - نودها در Nodes() تعریف شده‌اند
	return nil
}

// Unload خارج کردن از حافظه
func (p *CorePlugin) Unload() error {
	// Core plugin نمی‌تواند unload شود
	return nil
}

// Nodes لیست نودهای ارائه شده
func (p *CorePlugin) Nodes() []NodeDefinition {
	return []NodeDefinition{
		{
			Type:        "inject",
			Name:        "Inject",
			Category:    "core",
			Description: "ارسال پیام دوره‌ای یا دستی",
			Icon:        "play",
			Color:       "#3FADB5",
			Inputs:      0,
			Outputs:     1,
			Factory:     func() node.Executor { return coreNodes.NewInjectNode() },
			Config: map[string]interface{}{
				"payload":  "",
				"interval": 1000,
				"repeat":   true,
			},
		},
		{
			Type:        "debug",
			Name:        "Debug",
			Category:    "core",
			Description: "نمایش پیام در کنسول",
			Icon:        "bug",
			Color:       "#87A980",
			Inputs:      1,
			Outputs:     0,
			Factory:     func() node.Executor { return coreNodes.NewDebugNode() },
			Config: map[string]interface{}{
				"complete": false,
			},
		},
		{
			Type:        "function",
			Name:        "Function",
			Category:    "core",
			Description:   "اجرای کد JavaScript",
			Icon:        "code",
			Color:       "#E7E7AE",
			Inputs:      1,
			Outputs:     1,
			Factory:     func() node.Executor { return coreNodes.NewFunctionNode() },
			Config: map[string]interface{}{
				"func": "return msg;",
			},
		},
		{
			Type:        "if",
			Name:        "If",
			Category:    "core",
			Description: "مسیریابی شرطی",
			Icon:        "git-branch",
			Color:       "#C0DEED",
			Inputs:      1,
			Outputs:     2,
			Factory:     func() node.Executor { return coreNodes.NewIfNode() },
			Config: map[string]interface{}{
				"condition": "msg.payload > 0",
			},
		},
		{
			Type:        "delay",
			Name:        "Delay",
			Category:    "core",
			Description: "تأخیر در ارسال پیام",
			Icon:        "clock",
			Color:       "#C0C0C0",
			Inputs:      1,
			Outputs:     1,
			Factory:     func() node.Executor { return coreNodes.NewDelayNode() },
			Config: map[string]interface{}{
				"delay": 1000,
			},
		},
	}
}

// RequiredMemory حافظه مورد نیاز (MB)
func (p *CorePlugin) RequiredMemory() uint64 {
	return 5 * 1024 * 1024 // 5MB
}

// RequiredDisk فضای دیسک مورد نیاز
func (p *CorePlugin) RequiredDisk() uint64 {
	return 1 * 1024 * 1024 // 1MB
}

// Dependencies وابستگی‌ها
func (p *CorePlugin) Dependencies() []string {
	return []string{} // هیچ وابستگی ندارد
}
