package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/viper"
)

// Profile represents a build/runtime profile
type Profile string

const (
	// ProfileMinimal - Pi Zero, BeagleBone (512MB RAM)
	ProfileMinimal Profile = "minimal"

	// ProfileStandard - Pi 3/4, Orange Pi (1GB RAM)
	ProfileStandard Profile = "standard"

	// ProfileFull - Pi 4/5, Jetson Nano (2GB+ RAM)
	ProfileFull Profile = "full"
)

// ProfileConfig holds profile-specific configuration
type ProfileConfig struct {
	Name        Profile `mapstructure:"name"`
	Description string  `mapstructure:"description"`

	// Resource limits
	MaxMemory     int64 `mapstructure:"max_memory"`      // Max memory in MB
	MaxGoroutines int   `mapstructure:"max_goroutines"`  // Max concurrent goroutines
	MaxNodes      int   `mapstructure:"max_nodes"`       // Max nodes per flow
	MaxFlows      int   `mapstructure:"max_flows"`       // Max concurrent flows

	// Module configuration
	Modules ModulesConfig `mapstructure:"modules"`

	// Feature flags
	Features FeaturesConfig `mapstructure:"features"`
}

// ModulesConfig defines which modules are enabled for a profile
type ModulesConfig struct {
	Core       bool `mapstructure:"core"`       // Always enabled
	Network    bool `mapstructure:"network"`    // HTTP, MQTT, WebSocket, TCP, UDP
	GPIO       bool `mapstructure:"gpio"`       // GPIO, I2C, SPI, PWM
	Database   bool `mapstructure:"database"`   // MySQL, PostgreSQL, MongoDB, Redis
	Messaging  bool `mapstructure:"messaging"`  // Telegram, Email, Slack, Discord
	AI         bool `mapstructure:"ai"`         // OpenAI, Anthropic, Ollama
	Industrial bool `mapstructure:"industrial"` // Modbus, OPC-UA
	Parsers    bool `mapstructure:"parsers"`    // JSON, XML, CSV, YAML
}

// FeaturesConfig defines feature flags
type FeaturesConfig struct {
	WebUI           bool `mapstructure:"web_ui"`            // Enable web UI
	APIAuth         bool `mapstructure:"api_auth"`          // Enable API authentication
	Metrics         bool `mapstructure:"metrics"`           // Enable Prometheus metrics
	DebugMode       bool `mapstructure:"debug_mode"`        // Enable debug mode
	HotReload       bool `mapstructure:"hot_reload"`        // Enable flow hot reload
	AutoDisable     bool `mapstructure:"auto_disable"`      // Auto-disable modules on low memory
	ResourceMonitor bool `mapstructure:"resource_monitor"`  // Enable resource monitoring
}

// GetDefaultProfiles returns the default profile configurations
func GetDefaultProfiles() map[Profile]*ProfileConfig {
	return map[Profile]*ProfileConfig{
		ProfileMinimal: {
			Name:          ProfileMinimal,
			Description:   "Minimal profile for Pi Zero, BeagleBone (512MB RAM)",
			MaxMemory:     50,   // 50MB
			MaxGoroutines: 50,   // Limited goroutines
			MaxNodes:      50,   // Max 50 nodes per flow
			MaxFlows:      5,    // Max 5 concurrent flows
			Modules: ModulesConfig{
				Core:       true,
				Network:    false,
				GPIO:       false,
				Database:   false,
				Messaging:  false,
				AI:         false,
				Industrial: false,
				Parsers:    false,
			},
			Features: FeaturesConfig{
				WebUI:           false, // No web UI on minimal
				APIAuth:         false,
				Metrics:         false,
				DebugMode:       false,
				HotReload:       false,
				AutoDisable:     true,
				ResourceMonitor: true,
			},
		},
		ProfileStandard: {
			Name:          ProfileStandard,
			Description:   "Standard profile for Pi 3/4, Orange Pi (1GB RAM)",
			MaxMemory:     200,  // 200MB
			MaxGoroutines: 200,  // More goroutines
			MaxNodes:      500,  // Max 500 nodes per flow
			MaxFlows:      20,   // Max 20 concurrent flows
			Modules: ModulesConfig{
				Core:       true,
				Network:    true,
				GPIO:       true,
				Database:   true,
				Messaging:  false,
				AI:         false,
				Industrial: false,
				Parsers:    true,
			},
			Features: FeaturesConfig{
				WebUI:           true,
				APIAuth:         true,
				Metrics:         true,
				DebugMode:       false,
				HotReload:       true,
				AutoDisable:     true,
				ResourceMonitor: true,
			},
		},
		ProfileFull: {
			Name:          ProfileFull,
			Description:   "Full profile for Pi 4/5, Jetson Nano (2GB+ RAM)",
			MaxMemory:     400,  // 400MB
			MaxGoroutines: 1000, // Many goroutines
			MaxNodes:      1000, // Max 1000 nodes per flow
			MaxFlows:      100,  // Max 100 concurrent flows
			Modules: ModulesConfig{
				Core:       true,
				Network:    true,
				GPIO:       true,
				Database:   true,
				Messaging:  true,
				AI:         true,
				Industrial: true,
				Parsers:    true,
			},
			Features: FeaturesConfig{
				WebUI:           true,
				APIAuth:         true,
				Metrics:         true,
				DebugMode:       true,
				HotReload:       true,
				AutoDisable:     false,
				ResourceMonitor: true,
			},
		},
	}
}

// LoadProfile loads a profile configuration
func LoadProfile(profileName string) (*ProfileConfig, error) {
	profile := Profile(profileName)

	// Get default profiles
	defaults := GetDefaultProfiles()
	defaultConfig, exists := defaults[profile]
	if !exists {
		return nil, fmt.Errorf("unknown profile: %s", profileName)
	}

	// Try to load custom profile configuration
	v := viper.New()
	v.SetConfigName(fmt.Sprintf("profile-%s", profileName))
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath(getConfigDir())

	// Read profile config if exists
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read profile config: %w", err)
		}
		// Use defaults if no custom config
		return defaultConfig, nil
	}

	// Unmarshal into ProfileConfig
	var cfg ProfileConfig
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal profile config: %w", err)
	}

	// Merge with defaults (for any missing fields)
	mergeProfileConfig(&cfg, defaultConfig)

	return &cfg, nil
}

// DetectProfile automatically detects the best profile for the current system
func DetectProfile() Profile {
	// Check available memory
	var memInfo runtime.MemStats
	runtime.ReadMemStats(&memInfo)

	// Simple heuristic based on system memory
	// In production, we'd read /proc/meminfo on Linux
	totalMem := memInfo.Sys / 1024 / 1024 // Convert to MB

	// Check if running on ARM (likely Raspberry Pi)
	isARM := runtime.GOARCH == "arm" || runtime.GOARCH == "arm64"

	if !isARM {
		// Non-ARM systems get full profile
		return ProfileFull
	}

	// ARM-based systems
	if totalMem < 256 {
		return ProfileMinimal
	} else if totalMem < 1024 {
		return ProfileStandard
	}

	return ProfileFull
}

// DetectBoard attempts to detect the board type
func DetectBoard() string {
	// Check for Raspberry Pi
	if _, err := os.Stat("/proc/device-tree/model"); err == nil {
		data, err := os.ReadFile("/proc/device-tree/model")
		if err == nil {
			model := string(data)
			if contains(model, "Raspberry Pi Zero") {
				return "Pi Zero"
			} else if contains(model, "Raspberry Pi 3") {
				return "Pi 3"
			} else if contains(model, "Raspberry Pi 4") {
				return "Pi 4"
			} else if contains(model, "Raspberry Pi 5") {
				return "Pi 5"
			} else if contains(model, "Raspberry Pi") {
				return "Raspberry Pi"
			}
		}
	}

	// Check for BeagleBone
	if _, err := os.Stat("/etc/dogtag"); err == nil {
		return "BeagleBone"
	}

	// Check for Orange Pi
	if _, err := os.Stat("/etc/orangepi-release"); err == nil {
		return "Orange Pi"
	}

	// Check for Jetson
	if _, err := os.Stat("/etc/nv_tegra_release"); err == nil {
		return "Jetson"
	}

	// Generic Linux
	if runtime.GOOS == "linux" {
		if runtime.GOARCH == "arm64" {
			return "ARM64 Linux"
		} else if runtime.GOARCH == "arm" {
			return "ARM Linux"
		}
		return "Linux"
	}

	return "Unknown"
}

// GetProfileForBoard returns the recommended profile for a board type
func GetProfileForBoard(board string) Profile {
	switch board {
	case "Pi Zero":
		return ProfileMinimal
	case "Pi 3", "Orange Pi", "BeagleBone":
		return ProfileStandard
	case "Pi 4", "Pi 5", "Jetson":
		return ProfileFull
	default:
		return ProfileStandard
	}
}

// mergeProfileConfig merges two profile configs, using defaults for missing values
func mergeProfileConfig(cfg *ProfileConfig, defaults *ProfileConfig) {
	if cfg.Name == "" {
		cfg.Name = defaults.Name
	}
	if cfg.Description == "" {
		cfg.Description = defaults.Description
	}
	if cfg.MaxMemory == 0 {
		cfg.MaxMemory = defaults.MaxMemory
	}
	if cfg.MaxGoroutines == 0 {
		cfg.MaxGoroutines = defaults.MaxGoroutines
	}
	if cfg.MaxNodes == 0 {
		cfg.MaxNodes = defaults.MaxNodes
	}
	if cfg.MaxFlows == 0 {
		cfg.MaxFlows = defaults.MaxFlows
	}
}

// contains checks if a string contains a substring (case-insensitive helper)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SaveProfileConfig saves a profile configuration to file
func SaveProfileConfig(profileName string, cfg *ProfileConfig) error {
	configPath := filepath.Join(getConfigDir(), fmt.Sprintf("profile-%s.yaml", profileName))

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	v := viper.New()
	v.Set("name", cfg.Name)
	v.Set("description", cfg.Description)
	v.Set("max_memory", cfg.MaxMemory)
	v.Set("max_goroutines", cfg.MaxGoroutines)
	v.Set("max_nodes", cfg.MaxNodes)
	v.Set("max_flows", cfg.MaxFlows)
	v.Set("modules", cfg.Modules)
	v.Set("features", cfg.Features)

	return v.WriteConfigAs(configPath)
}

// ValidateProfile validates a profile configuration
func ValidateProfile(cfg *ProfileConfig) error {
	if cfg.MaxMemory < 10 {
		return fmt.Errorf("max_memory must be at least 10MB")
	}
	if cfg.MaxGoroutines < 10 {
		return fmt.Errorf("max_goroutines must be at least 10")
	}
	if cfg.MaxNodes < 1 {
		return fmt.Errorf("max_nodes must be at least 1")
	}
	if cfg.MaxFlows < 1 {
		return fmt.Errorf("max_flows must be at least 1")
	}
	return nil
}
