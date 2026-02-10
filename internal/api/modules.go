package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/module/manager"
	"github.com/edgeflow/edgeflow/internal/module/parser"
	"github.com/edgeflow/edgeflow/internal/module/validator"
	"github.com/gofiber/fiber/v2"
)

// HTTP client with timeout for module downloads
var httpClient = &http.Client{
	Timeout: 5 * time.Minute,
}

// ModuleAPI handles module management endpoints
type ModuleAPI struct {
	manager   *manager.ModuleManager
	validator *validator.Validator
	uploadDir string
}

// NewModuleAPI creates a new module API handler
func NewModuleAPI(modulesDir string) (*ModuleAPI, error) {
	uploadDir := filepath.Join(modulesDir, ".uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		// Use temp dir as fallback
		uploadDir = os.TempDir()
	}

	api := &ModuleAPI{
		validator: validator.NewValidator(),
		uploadDir: uploadDir,
	}

	// Try to initialize module manager, but don't fail if it doesn't work
	// This allows search functionality to work even without full module management
	mgr, err := manager.NewModuleManager(modulesDir)
	if err == nil {
		api.manager = mgr
	}

	return api, nil
}

// NewModuleAPILegacy creates a new module API handler (legacy behavior)
func NewModuleAPILegacy(modulesDir string) (*ModuleAPI, error) {
	mgr, err := manager.NewModuleManager(modulesDir)
	if err != nil {
		return nil, err
	}

	uploadDir := filepath.Join(modulesDir, ".uploads")
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	return &ModuleAPI{
		manager:   mgr,
		validator: validator.NewValidator(),
		uploadDir: uploadDir,
	}, nil
}

// SetupModuleRoutes configures module API routes
func SetupModuleRoutes(app *fiber.App, api *ModuleAPI) {
	modules := app.Group("/api/v1/modules")

	// List all modules
	modules.Get("/", api.ListModules)

	// Search endpoints (must be before /:name to avoid conflicts)
	modules.Get("/search/npm", api.SearchNPM)
	modules.Get("/search/nodered", api.SearchNodeRED)
	modules.Get("/search/github", api.SearchGitHub)

	// Get module details
	modules.Get("/:name", api.GetModule)

	// Install module from URL or upload
	modules.Post("/install", api.InstallModule)

	// Upload and install module
	modules.Post("/upload", api.UploadModule)

	// Validate module without installing
	modules.Post("/validate", api.ValidateModule)

	// Uninstall module
	modules.Delete("/:name", api.UninstallModule)

	// Enable/disable module
	modules.Post("/:name/enable", api.EnableModule)
	modules.Post("/:name/disable", api.DisableModule)

	// Load/unload module
	modules.Post("/:name/load", api.LoadModule)
	modules.Post("/:name/unload", api.UnloadModule)

	// Get module nodes
	modules.Get("/:name/nodes", api.GetModuleNodes)

	// License endpoints
	modules.Get("/:name/license", api.GetModuleLicense)
	modules.Get("/licenses", api.GetAllLicenses)
	modules.Post("/license/check", api.CheckLicenseCompatibility)

	// Parse module info (preview without install)
	modules.Post("/parse", api.ParseModule)
}

// ListModules returns all installed modules
func (api *ModuleAPI) ListModules(c *fiber.Ctx) error {
	if api.manager == nil {
		return c.JSON(fiber.Map{
			"modules": []interface{}{},
			"count":   0,
		})
	}

	modules := api.manager.List()

	result := make([]fiber.Map, 0, len(modules))
	for _, mod := range modules {
		// Map manager status to frontend status
		status := mapModuleStatus(mod.Status, mod.Enabled)

		// Build nodes array for frontend
		nodes := make([]fiber.Map, 0, len(mod.Info.Nodes))
		for _, n := range mod.Info.Nodes {
			nodes = append(nodes, fiber.Map{
				"type":        n.Type,
				"name":        n.Name,
				"category":    n.Category,
				"description": n.Description,
				"icon":        n.Icon,
				"color":       n.Color,
				"inputs":      n.Inputs,
				"outputs":     n.Outputs,
			})
		}

		// Determine category from the first node, or default
		category := "advanced"
		if len(mod.Info.Nodes) > 0 && mod.Info.Nodes[0].Category != "" {
			category = mod.Info.Nodes[0].Category
		}

		// Dependencies from keywords/config (modules don't have explicit deps yet)
		dependencies := make([]string, 0)

		result = append(result, fiber.Map{
			"name":               mod.Info.Name,
			"version":            mod.Info.Version,
			"description":        mod.Info.Description,
			"author":             mod.Info.Author,
			"category":           category,
			"status":             status,
			"loaded_at":          mod.InstalledAt,
			"error":              mod.Error,
			"required_memory_mb": 0,
			"required_disk_mb":   0,
			"dependencies":       dependencies,
			"nodes":              nodes,
			"config":             mod.Info.Config,
			"compatible":         true,
			"compatible_reason":  "",
		})
	}

	return c.JSON(fiber.Map{
		"modules": result,
		"count":   len(result),
	})
}

// mapModuleStatus converts manager.ModuleStatus to frontend status strings
func mapModuleStatus(status manager.ModuleStatus, enabled bool) string {
	switch status {
	case manager.StatusLoaded:
		return "loaded"
	case manager.StatusError:
		return "error"
	case manager.StatusDisabled:
		return "not_loaded"
	case manager.StatusInstalled, manager.StatusPending:
		return "not_loaded"
	default:
		return "not_loaded"
	}
}

// ReloadModule unloads and reloads a module
func (api *ModuleAPI) ReloadModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if api.manager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Module manager not available",
		})
	}

	// Unload then load
	if err := api.manager.Unload(name); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Reload failed (unload): %v", err),
		})
	}

	if err := api.manager.Load(name); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Reload failed (load): %v", err),
		})
	}

	mod, _ := api.manager.Get(name)
	return c.JSON(fiber.Map{
		"message":      "Module reloaded",
		"loaded_nodes": mod.LoadedNodes,
	})
}

// GetModuleStats returns module statistics
func (api *ModuleAPI) GetModuleStats(c *fiber.Ctx) error {
	if api.manager == nil {
		return c.JSON(fiber.Map{
			"total_plugins":       0,
			"loaded_plugins":      0,
			"enabled_plugins":     0,
			"total_nodes":         0,
			"load_order":          []string{},
			"memory_available_mb": 0,
		})
	}

	modules := api.manager.List()

	totalPlugins := len(modules)
	loadedPlugins := 0
	enabledPlugins := 0
	totalNodes := 0
	loadOrder := make([]string, 0)

	for _, mod := range modules {
		if mod.Status == manager.StatusLoaded {
			loadedPlugins++
			loadOrder = append(loadOrder, mod.Info.Name)
		}
		if mod.Enabled {
			enabledPlugins++
		}
		totalNodes += len(mod.Info.Nodes)
	}

	return c.JSON(fiber.Map{
		"total_plugins":       totalPlugins,
		"loaded_plugins":      loadedPlugins,
		"enabled_plugins":     enabledPlugins,
		"total_nodes":         totalNodes,
		"load_order":          loadOrder,
		"memory_available_mb": 0,
	})
}

// GetModule returns details of a specific module
func (api *ModuleAPI) GetModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if api.manager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Module manager not available",
		})
	}

	mod, ok := api.manager.Get(name)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Module not found",
		})
	}

	status := mapModuleStatus(mod.Status, mod.Enabled)

	// Build nodes array
	nodes := make([]fiber.Map, 0, len(mod.Info.Nodes))
	for _, n := range mod.Info.Nodes {
		nodes = append(nodes, fiber.Map{
			"type":        n.Type,
			"name":        n.Name,
			"category":    n.Category,
			"description": n.Description,
			"icon":        n.Icon,
			"color":       n.Color,
			"inputs":      n.Inputs,
			"outputs":     n.Outputs,
		})
	}

	category := "advanced"
	if len(mod.Info.Nodes) > 0 && mod.Info.Nodes[0].Category != "" {
		category = mod.Info.Nodes[0].Category
	}

	return c.JSON(fiber.Map{
		"name":               mod.Info.Name,
		"version":            mod.Info.Version,
		"description":        mod.Info.Description,
		"author":             mod.Info.Author,
		"category":           category,
		"status":             status,
		"loaded_at":          mod.InstalledAt,
		"error":              mod.Error,
		"required_memory_mb": 0,
		"required_disk_mb":   0,
		"dependencies":       []string{},
		"nodes":              nodes,
		"config":             mod.Info.Config,
		"compatible":         true,
		"compatible_reason":  "",
		"validation":         mod.Validation,
		"loaded_nodes":       mod.LoadedNodes,
	})
}

// InstallRequest represents module installation request
type InstallRequest struct {
	URL    string `json:"url"`
	Path   string `json:"path"`
	NPM    string `json:"npm"`    // npm package name
	GitHub string `json:"github"` // GitHub repo (owner/repo)
}

// InstallModule installs a module from URL or path
func (api *ModuleAPI) InstallModule(c *fiber.Ctx) error {
	var req InstallRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var sourcePath string
	var err error

	switch {
	case req.Path != "":
		sourcePath = req.Path
	case req.URL != "":
		sourcePath, err = downloadModule(req.URL, api.uploadDir)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to download module: %v", err),
			})
		}
		defer os.RemoveAll(sourcePath)
	case req.NPM != "":
		sourcePath, err = downloadNPMPackage(req.NPM, api.uploadDir)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to download npm package: %v", err),
			})
		}
		defer os.RemoveAll(sourcePath)
	case req.GitHub != "":
		sourcePath, err = downloadGitHubRepo(req.GitHub, api.uploadDir)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to download GitHub repo: %v", err),
			})
		}
		defer os.RemoveAll(sourcePath)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Must provide url, path, npm, or github",
		})
	}

	// Install module
	installed, err := api.manager.Install(sourcePath)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Installation failed: %v", err),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Module installed successfully",
		"module":  installed.Info,
		"status":  installed.Status,
		"validation": installed.Validation,
	})
}

// UploadModule handles module file upload and installation
func (api *ModuleAPI) UploadModule(c *fiber.Ctx) error {
	file, err := c.FormFile("module")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No module file provided",
		})
	}

	// Save uploaded file
	filename := filepath.Join(api.uploadDir, file.Filename)
	if err := c.SaveFile(file, filename); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to save file: %v", err),
		})
	}
	defer os.Remove(filename)

	// Install module
	installed, err := api.manager.Install(filename)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Installation failed: %v", err),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Module uploaded and installed successfully",
		"module":  installed.Info,
		"status":  installed.Status,
		"validation": installed.Validation,
	})
}

// ValidateModule validates a module without installing
func (api *ModuleAPI) ValidateModule(c *fiber.Ctx) error {
	file, err := c.FormFile("module")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No module file provided",
		})
	}

	// Save uploaded file temporarily
	filename := filepath.Join(api.uploadDir, file.Filename)
	if err := c.SaveFile(file, filename); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to save file: %v", err),
		})
	}
	defer os.Remove(filename)

	// Parse module
	format, err := parser.DetectFormat(filename)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to detect format: %v", err),
		})
	}

	var moduleParser parser.Parser
	switch format {
	case parser.FormatEdgeFlow:
		moduleParser = parser.NewEdgeFlowParser()
	case parser.FormatNodeRED:
		moduleParser = parser.NewNodeRedParser()
	case parser.FormatN8N:
		moduleParser = parser.NewN8NParser()
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Unsupported module format",
		})
	}

	info, err := moduleParser.Parse(filename)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to parse module: %v", err),
		})
	}

	// Validate
	result := api.validator.Validate(info)

	return c.JSON(fiber.Map{
		"module":     info,
		"validation": result,
	})
}

// ParseModule parses module info without installing
func (api *ModuleAPI) ParseModule(c *fiber.Ctx) error {
	file, err := c.FormFile("module")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No module file provided",
		})
	}

	// Save uploaded file temporarily
	filename := filepath.Join(api.uploadDir, file.Filename)
	if err := c.SaveFile(file, filename); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to save file: %v", err),
		})
	}
	defer os.Remove(filename)

	// Detect format
	format, err := parser.DetectFormat(filename)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to detect format: %v", err),
		})
	}

	var moduleParser parser.Parser
	switch format {
	case parser.FormatNodeRED:
		moduleParser = parser.NewNodeRedParser()
	case parser.FormatN8N:
		moduleParser = parser.NewN8NParser()
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":  "Unsupported module format",
			"format": format,
		})
	}

	info, err := moduleParser.Parse(filename)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to parse module: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"module": info,
		"format": format,
	})
}

// UninstallModule removes a module
func (api *ModuleAPI) UninstallModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if err := api.manager.Uninstall(name); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Uninstall failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Module uninstalled successfully",
	})
}

// EnableModule enables a module
func (api *ModuleAPI) EnableModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if api.manager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Module manager not available",
		})
	}

	if err := api.manager.Enable(name); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Enable failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Module enabled",
	})
}

// DisableModule disables a module
func (api *ModuleAPI) DisableModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if api.manager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Module manager not available",
		})
	}

	if err := api.manager.Disable(name); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Disable failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Module disabled",
	})
}

// LoadModule loads a module into the runtime
func (api *ModuleAPI) LoadModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if api.manager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Module manager not available",
		})
	}

	if err := api.manager.Load(name); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Load failed: %v", err),
		})
	}

	mod, _ := api.manager.Get(name)
	return c.JSON(fiber.Map{
		"message":      "Module loaded",
		"loaded_nodes": mod.LoadedNodes,
	})
}

// UnloadModule unloads a module from the runtime
func (api *ModuleAPI) UnloadModule(c *fiber.Ctx) error {
	name := c.Params("name")

	if api.manager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Module manager not available",
		})
	}

	if err := api.manager.Unload(name); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Unload failed: %v", err),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Module unloaded",
	})
}

// GetModuleNodes returns nodes provided by a module
func (api *ModuleAPI) GetModuleNodes(c *fiber.Ctx) error {
	name := c.Params("name")

	mod, ok := api.manager.Get(name)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Module not found",
		})
	}

	return c.JSON(fiber.Map{
		"nodes": mod.Info.Nodes,
		"total": len(mod.Info.Nodes),
	})
}

// GetModuleLicense returns license information for a module
func (api *ModuleAPI) GetModuleLicense(c *fiber.Ctx) error {
	name := c.Params("name")

	mod, ok := api.manager.Get(name)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Module not found",
		})
	}

	return c.JSON(fiber.Map{
		"module":      name,
		"license":     mod.Info.License,
		"license_info": mod.LicenseInfo,
		"attribution": mod.Attribution,
	})
}

// GetAllLicenses returns license information for all installed modules
func (api *ModuleAPI) GetAllLicenses(c *fiber.Ctx) error {
	if api.manager == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Module manager not available",
		})
	}

	attributions := api.manager.GetAllAttributions()

	// Build license summary
	licenses := make(map[string][]string)
	for _, attr := range attributions {
		if attr.License != "" {
			licenses[attr.License] = append(licenses[attr.License], attr.ModuleName)
		}
	}

	return c.JSON(fiber.Map{
		"total_modules": len(attributions),
		"attributions":  attributions,
		"by_license":    licenses,
	})
}

// CheckLicenseCompatibility checks if a package license is compatible before install
func (api *ModuleAPI) CheckLicenseCompatibility(c *fiber.Ctx) error {
	var req struct {
		License string `json:"license"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	licenseInfo := validator.AnalyzeLicense(req.License)

	return c.JSON(fiber.Map{
		"license":       req.License,
		"license_info":  licenseInfo,
		"compatible":    licenseInfo.Compatibility == "compatible",
		"warning":       licenseInfo.Compatibility == "warning",
		"incompatible":  licenseInfo.Compatibility == "incompatible",
	})
}

// ============================================
// Download Functions
// ============================================

// downloadModule downloads a module from a URL
func downloadModule(moduleURL string, destDir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", moduleURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "EdgeFlow/1.0 Module Downloader")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	// Determine file extension
	ext := determineExtension(moduleURL, resp.Header.Get("Content-Type"))

	// Save to temp file
	tmpPath, err := CopyToTemp(resp.Body, destDir, "module"+ext)
	if err != nil {
		return "", fmt.Errorf("failed to save module: %w", err)
	}

	return tmpPath, nil
}

func determineExtension(moduleURL string, contentType string) string {
	if strings.HasSuffix(moduleURL, ".zip") {
		return ".zip"
	}
	if strings.HasSuffix(moduleURL, ".tgz") || strings.HasSuffix(moduleURL, ".tar.gz") {
		return ".tgz"
	}
	switch contentType {
	case "application/zip":
		return ".zip"
	case "application/gzip", "application/x-gzip":
		return ".tgz"
	default:
		return ".tgz"
	}
}

// NPMPackageInfo represents npm registry response
type NPMPackageInfo struct {
	Name     string                    `json:"name"`
	DistTags map[string]string         `json:"dist-tags"`
	Versions map[string]NPMVersionInfo `json:"versions"`
}

type NPMVersionInfo struct {
	Version string      `json:"version"`
	Dist    NPMDistInfo `json:"dist"`
}

type NPMDistInfo struct {
	Tarball   string `json:"tarball"`
	Shasum    string `json:"shasum"`
	Integrity string `json:"integrity"`
}

// downloadNPMPackage downloads a package from npm registry
func downloadNPMPackage(packageName string, destDir string) (string, error) {
	name, version := parseNPMPackageName(packageName)

	// Fetch package metadata
	registryURL := fmt.Sprintf("https://registry.npmjs.org/%s", url.PathEscape(name))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", registryURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create registry request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch package info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("package not found: %s", packageName)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("registry error: %d", resp.StatusCode)
	}

	var pkgInfo NPMPackageInfo
	if err := json.NewDecoder(resp.Body).Decode(&pkgInfo); err != nil {
		return "", fmt.Errorf("failed to parse package info: %w", err)
	}

	// Get version
	if version == "" {
		version = pkgInfo.DistTags["latest"]
	}

	versionInfo, ok := pkgInfo.Versions[version]
	if !ok {
		return "", fmt.Errorf("version not found: %s@%s", name, version)
	}

	// Download tarball
	return downloadModule(versionInfo.Dist.Tarball, destDir)
}

func parseNPMPackageName(packageName string) (name, version string) {
	// Handle scoped packages (@scope/name@version)
	if strings.HasPrefix(packageName, "@") {
		parts := strings.SplitN(packageName[1:], "@", 2)
		if len(parts) == 2 {
			return "@" + parts[0], parts[1]
		}
		return packageName, ""
	}
	// Handle regular packages (name@version)
	parts := strings.SplitN(packageName, "@", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return packageName, ""
}

// downloadGitHubRepo downloads a repository from GitHub
func downloadGitHubRepo(repo string, destDir string) (string, error) {
	owner, repoName, branch := parseGitHubRepo(repo)
	if owner == "" || repoName == "" {
		return "", fmt.Errorf("invalid GitHub repo format, expected owner/repo[@branch]")
	}

	if branch == "" {
		branch = "HEAD"
	}

	// Build download URL
	downloadURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/zipball/%s", owner, repoName, branch)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create GitHub request: %w", err)
	}

	req.Header.Set("User-Agent", "EdgeFlow/1.0 Module Downloader")
	req.Header.Set("Accept", "application/vnd.github+json")

	// Add token if available
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download from GitHub: %w", err)
	}
	defer resp.Body.Close()

	// Check rate limiting
	if resp.Header.Get("X-RateLimit-Remaining") == "0" {
		return "", fmt.Errorf("GitHub rate limit exceeded")
	}

	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("repository not found: %s", repo)
	}
	if resp.StatusCode == http.StatusForbidden {
		return "", fmt.Errorf("GitHub rate limit exceeded or access denied")
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub error: %d", resp.StatusCode)
	}

	// Save to temp file
	tmpPath, err := CopyToTemp(resp.Body, destDir, repoName+"_.zip")
	if err != nil {
		return "", fmt.Errorf("failed to save GitHub repo: %w", err)
	}

	return tmpPath, nil
}

func parseGitHubRepo(repo string) (owner, name, branch string) {
	// Handle owner/repo@branch format
	if idx := strings.Index(repo, "@"); idx > 0 {
		branch = repo[idx+1:]
		repo = repo[:idx]
	}

	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", ""
	}

	return parts[0], parts[1], branch
}

// CopyToTemp copies an io.Reader to a temp file, preserving the file extension
func CopyToTemp(r io.Reader, destDir, filename string) (string, error) {
	ext := filepath.Ext(filename)
	base := strings.TrimSuffix(filename, ext)
	// Place the wildcard before the extension so os.CreateTemp preserves it
	// e.g. "module_*.tgz" produces "module_1234567890.tgz"
	tmpFile, err := os.CreateTemp(destDir, base+"_*"+ext)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, r); err != nil {
		os.Remove(tmpFile.Name())
		return "", err
	}

	return tmpFile.Name(), nil
}

// ============================================
// Search API Handlers
// ============================================

// NPMSearchResult represents npm search API response
type NPMSearchResult struct {
	Objects []NPMSearchObject `json:"objects"`
	Total   int               `json:"total"`
}

type NPMSearchObject struct {
	Package NPMSearchPackage `json:"package"`
	Score   NPMSearchScore   `json:"score"`
}

type NPMSearchPackage struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Keywords    []string `json:"keywords"`
	Publisher   struct {
		Username string `json:"username"`
	} `json:"publisher"`
	Links struct {
		NPM        string `json:"npm"`
		Homepage   string `json:"homepage"`
		Repository string `json:"repository"`
	} `json:"links"`
}

type NPMSearchScore struct {
	Final float64 `json:"final"`
}

// SearchNPM searches npm registry for Node-RED and n8n modules
func (api *ModuleAPI) SearchNPM(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	// Search npm with "node-red" prefix to get Node-RED packages
	searchURL := fmt.Sprintf(
		"https://registry.npmjs.org/-/v1/search?text=node-red+%s&size=100",
		url.QueryEscape(query),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create search request",
		})
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Failed to reach npm registry",
		})
	}
	defer resp.Body.Close()

	var result NPMSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse npm response",
		})
	}

	// Return all results - they're already filtered by "node-red" in search query
	results := make([]fiber.Map, 0, len(result.Objects))
	for _, obj := range result.Objects {
		results = append(results, fiber.Map{
			"name":        obj.Package.Name,
			"version":     obj.Package.Version,
			"description": obj.Package.Description,
			"keywords":    obj.Package.Keywords,
			"author":      obj.Package.Publisher.Username,
			"url":         obj.Package.Links.NPM,
			"repository":  obj.Package.Links.Repository,
			"score":       obj.Score.Final,
			"source":      "npm",
		})
	}

	return c.JSON(fiber.Map{
		"results": results,
		"total":   len(results),
		"query":   query,
	})
}

// Node-RED catalog cache
var (
	nodeRedCatalogCache     []NodeREDCatalogModule
	nodeRedCatalogCacheTime time.Time
	nodeRedCatalogMutex     sync.RWMutex
)

// NodeREDCatalogModule represents a module in Node-RED catalog
type NodeREDCatalogModule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Version     string   `json:"version"`
	Keywords    []string `json:"keywords"`
	Updated     string   `json:"updated_at"`
	Types       []string `json:"types"`
}

// SearchNodeRED searches the Node-RED catalog
func (api *ModuleAPI) SearchNodeRED(c *fiber.Ctx) error {
	query := strings.ToLower(c.Query("q"))
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	catalog, err := getNodeREDCatalog()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch Node-RED catalog: %v", err),
		})
	}

	// Search through catalog
	results := make([]fiber.Map, 0)
	for _, mod := range catalog {
		if matchesNodeREDQuery(mod, query) {
			results = append(results, fiber.Map{
				"name":        mod.Name,
				"version":     mod.Version,
				"description": mod.Description,
				"keywords":    mod.Keywords,
				"types":       mod.Types,
				"updated":     mod.Updated,
				"source":      "node-red",
				"url":         fmt.Sprintf("https://flows.nodered.org/node/%s", mod.Name),
			})
		}
		if len(results) >= 50 {
			break
		}
	}

	return c.JSON(fiber.Map{
		"results": results,
		"total":   len(results),
		"query":   query,
	})
}

func getNodeREDCatalog() ([]NodeREDCatalogModule, error) {
	nodeRedCatalogMutex.RLock()
	if time.Since(nodeRedCatalogCacheTime) < time.Hour && nodeRedCatalogCache != nil {
		defer nodeRedCatalogMutex.RUnlock()
		return nodeRedCatalogCache, nil
	}
	nodeRedCatalogMutex.RUnlock()

	nodeRedCatalogMutex.Lock()
	defer nodeRedCatalogMutex.Unlock()

	// Double-check after acquiring write lock
	if time.Since(nodeRedCatalogCacheTime) < time.Hour && nodeRedCatalogCache != nil {
		return nodeRedCatalogCache, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET",
		"https://catalogue.nodered.org/catalogue.json", nil)
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var catalog struct {
		Modules []NodeREDCatalogModule `json:"modules"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return nil, err
	}

	nodeRedCatalogCache = catalog.Modules
	nodeRedCatalogCacheTime = time.Now()

	return nodeRedCatalogCache, nil
}

func matchesNodeREDQuery(mod NodeREDCatalogModule, query string) bool {
	if strings.Contains(strings.ToLower(mod.Name), query) {
		return true
	}
	if strings.Contains(strings.ToLower(mod.Description), query) {
		return true
	}
	for _, kw := range mod.Keywords {
		if strings.Contains(strings.ToLower(kw), query) {
			return true
		}
	}
	return false
}

// GitHubSearchResult represents GitHub search API response
type GitHubSearchResult struct {
	TotalCount int                `json:"total_count"`
	Items      []GitHubRepository `json:"items"`
}

type GitHubRepository struct {
	FullName    string   `json:"full_name"`
	Description string   `json:"description"`
	HTMLURL     string   `json:"html_url"`
	Stars       int      `json:"stargazers_count"`
	Forks       int      `json:"forks_count"`
	Language    string   `json:"language"`
	Topics      []string `json:"topics"`
	UpdatedAt   string   `json:"updated_at"`
	Owner       struct {
		Login     string `json:"login"`
		AvatarURL string `json:"avatar_url"`
	} `json:"owner"`
}

// SearchGitHub searches GitHub for module repositories
func (api *ModuleAPI) SearchGitHub(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	// Build GitHub search query focusing on node-red and n8n
	searchQuery := fmt.Sprintf("%s node-red OR n8n", query)
	searchURL := fmt.Sprintf(
		"https://api.github.com/search/repositories?q=%s&sort=stars&order=desc&per_page=30",
		url.QueryEscape(searchQuery),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create search request",
		})
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "EdgeFlow/1.0")

	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Failed to reach GitHub API",
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
			"error": "GitHub rate limit exceeded",
		})
	}

	var result GitHubSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to parse GitHub response",
		})
	}

	// Transform results
	results := make([]fiber.Map, 0, len(result.Items))
	for _, repo := range result.Items {
		results = append(results, fiber.Map{
			"name":        repo.FullName,
			"description": repo.Description,
			"url":         repo.HTMLURL,
			"stars":       repo.Stars,
			"forks":       repo.Forks,
			"language":    repo.Language,
			"topics":      repo.Topics,
			"updated":     repo.UpdatedAt,
			"owner":       repo.Owner.Login,
			"avatar":      repo.Owner.AvatarURL,
			"source":      "github",
		})
	}

	return c.JSON(fiber.Map{
		"results": results,
		"total":   result.TotalCount,
		"query":   query,
	})
}
