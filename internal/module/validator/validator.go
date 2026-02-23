// Package validator provides validation for imported modules
// Ensures security and compatibility before modules are loaded
package validator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/EdgxCloud/EdgeFlow/internal/module/parser"
)

// ValidationResult contains validation results
type ValidationResult struct {
	Valid       bool              `json:"valid"`
	Errors      []ValidationError `json:"errors,omitempty"`
	Warnings    []ValidationError `json:"warnings,omitempty"`
	Score       int               `json:"score"` // 0-100 safety score
	LicenseInfo *LicenseInfo      `json:"license_info,omitempty"`
}

// LicenseInfo contains license analysis results
type LicenseInfo struct {
	License          string   `json:"license"`
	LicenseType      string   `json:"license_type"`      // "permissive", "copyleft", "proprietary", "unknown"
	IsOSIApproved    bool     `json:"is_osi_approved"`
	CommercialUse    bool     `json:"commercial_use"`
	Modification     bool     `json:"modification"`
	Distribution     bool     `json:"distribution"`
	PatentGrant      bool     `json:"patent_grant"`
	Attribution      bool     `json:"attribution_required"`
	ShareAlike       bool     `json:"share_alike"`        // Copyleft requirement
	Compatibility    string   `json:"compatibility"`      // "compatible", "warning", "incompatible"
	CompatibilityMsg string   `json:"compatibility_msg"`
	Permissions      []string `json:"permissions,omitempty"`
	Conditions       []string `json:"conditions,omitempty"`
	Limitations      []string `json:"limitations,omitempty"`
}

// ValidationError represents a validation issue
type ValidationError struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	File     string `json:"file,omitempty"`
	Line     int    `json:"line,omitempty"`
	Severity string `json:"severity"` // "error", "warning", "info"
}

// Validator validates imported modules
type Validator struct {
	allowedModules    map[string]bool
	blockedPatterns   []*regexp.Regexp
	maxFileSize       int64
	maxTotalSize      int64
	allowNetworkAccess bool
	allowFileSystem   bool
	allowChildProcess bool
}

// NewValidator creates a new validator with default settings
func NewValidator() *Validator {
	return &Validator{
		allowedModules:    make(map[string]bool),
		blockedPatterns:   compileBlockedPatterns(),
		maxFileSize:       10 * 1024 * 1024,  // 10MB per file
		maxTotalSize:      50 * 1024 * 1024,  // 50MB total
		allowNetworkAccess: false,
		allowFileSystem:   false,
		allowChildProcess: false,
	}
}

// ValidatorOption configures the validator
type ValidatorOption func(*Validator)

// WithNetworkAccess allows network access
func WithNetworkAccess(allow bool) ValidatorOption {
	return func(v *Validator) {
		v.allowNetworkAccess = allow
	}
}

// WithFileSystemAccess allows file system access
func WithFileSystemAccess(allow bool) ValidatorOption {
	return func(v *Validator) {
		v.allowFileSystem = allow
	}
}

// WithChildProcess allows child process spawning
func WithChildProcess(allow bool) ValidatorOption {
	return func(v *Validator) {
		v.allowChildProcess = allow
	}
}

// WithMaxFileSize sets maximum file size
func WithMaxFileSize(size int64) ValidatorOption {
	return func(v *Validator) {
		v.maxFileSize = size
	}
}

// WithAllowedModule adds an allowed module
func WithAllowedModule(name string) ValidatorOption {
	return func(v *Validator) {
		v.allowedModules[name] = true
	}
}

// NewValidatorWithOptions creates a validator with options
func NewValidatorWithOptions(opts ...ValidatorOption) *Validator {
	v := NewValidator()
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// Validate validates a parsed module
func (v *Validator) Validate(moduleInfo *parser.ModuleInfo) *ValidationResult {
	result := &ValidationResult{
		Valid:    true,
		Errors:   []ValidationError{},
		Warnings: []ValidationError{},
		Score:    100,
	}

	// Validate metadata
	v.validateMetadata(moduleInfo, result)

	// Validate license
	v.validateLicense(moduleInfo, result)

	// Validate structure
	v.validateStructure(moduleInfo, result)

	// Validate nodes
	for _, nodeInfo := range moduleInfo.Nodes {
		v.validateNode(&nodeInfo, moduleInfo.SourcePath, result)
	}

	// Scan for security issues
	v.scanForSecurityIssues(moduleInfo.SourcePath, result)

	// Calculate final validity
	result.Valid = len(result.Errors) == 0

	return result
}

// validateMetadata validates module metadata
func (v *Validator) validateMetadata(info *parser.ModuleInfo, result *ValidationResult) {
	if info.Name == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "MISSING_NAME",
			Message:  "Module name is required",
			Severity: "error",
		})
		result.Score -= 20
	}

	if info.Version == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Code:     "MISSING_VERSION",
			Message:  "Module version is recommended",
			Severity: "warning",
		})
		result.Score -= 5
	}

	// Validate name format
	if info.Name != "" {
		if !isValidModuleName(info.Name) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "INVALID_NAME",
				Message:  fmt.Sprintf("Invalid module name: %s", info.Name),
				Severity: "error",
			})
			result.Score -= 10
		}
	}

	// Check for description
	if info.Description == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Code:     "MISSING_DESCRIPTION",
			Message:  "Module description is recommended",
			Severity: "warning",
		})
		result.Score -= 2
	}

	// License is now validated separately in validateLicense
}

// validateLicense validates module license and checks compatibility
func (v *Validator) validateLicense(info *parser.ModuleInfo, result *ValidationResult) {
	licenseInfo := AnalyzeLicense(info.License)
	result.LicenseInfo = licenseInfo

	// No license specified
	if info.License == "" {
		result.Warnings = append(result.Warnings, ValidationError{
			Code:     "MISSING_LICENSE",
			Message:  "Module license is not specified. Usage terms are unclear.",
			Severity: "warning",
		})
		result.Score -= 5
		return
	}

	// Check license compatibility
	switch licenseInfo.Compatibility {
	case "incompatible":
		result.Errors = append(result.Errors, ValidationError{
			Code:     "INCOMPATIBLE_LICENSE",
			Message:  licenseInfo.CompatibilityMsg,
			Severity: "error",
		})
		result.Score -= 30

	case "warning":
		result.Warnings = append(result.Warnings, ValidationError{
			Code:     "LICENSE_WARNING",
			Message:  licenseInfo.CompatibilityMsg,
			Severity: "warning",
		})
		result.Score -= 10

	case "unknown":
		result.Warnings = append(result.Warnings, ValidationError{
			Code:     "UNKNOWN_LICENSE",
			Message:  fmt.Sprintf("Unknown license '%s'. Please review terms manually before using.", info.License),
			Severity: "warning",
		})
		result.Score -= 5
	}

	// Warn about copyleft licenses
	if licenseInfo.ShareAlike {
		result.Warnings = append(result.Warnings, ValidationError{
			Code:     "COPYLEFT_LICENSE",
			Message:  fmt.Sprintf("License '%s' is copyleft. Derivative works must use the same license.", info.License),
			Severity: "warning",
		})
	}

	// Warn about attribution requirements
	if licenseInfo.Attribution {
		result.Warnings = append(result.Warnings, ValidationError{
			Code:     "ATTRIBUTION_REQUIRED",
			Message:  fmt.Sprintf("License '%s' requires attribution. Include copyright notice in your project.", info.License),
			Severity: "info",
		})
	}
}

// AnalyzeLicense analyzes a license string and returns detailed info
func AnalyzeLicense(license string) *LicenseInfo {
	license = strings.TrimSpace(license)
	upperLicense := strings.ToUpper(license)

	// Default unknown license
	info := &LicenseInfo{
		License:       license,
		LicenseType:   "unknown",
		Compatibility: "unknown",
	}

	if license == "" {
		info.LicenseType = "unspecified"
		info.Compatibility = "unknown"
		info.CompatibilityMsg = "No license specified"
		return info
	}

	// Permissive licenses - fully compatible
	permissiveLicenses := map[string]*LicenseInfo{
		"MIT": {
			License:       "MIT",
			LicenseType:   "permissive",
			IsOSIApproved: true,
			CommercialUse: true,
			Modification:  true,
			Distribution:  true,
			PatentGrant:   false,
			Attribution:   true,
			ShareAlike:    false,
			Compatibility: "compatible",
			Permissions:   []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:    []string{"License and copyright notice"},
			Limitations:   []string{"No liability", "No warranty"},
		},
		"ISC": {
			License:       "ISC",
			LicenseType:   "permissive",
			IsOSIApproved: true,
			CommercialUse: true,
			Modification:  true,
			Distribution:  true,
			PatentGrant:   false,
			Attribution:   true,
			ShareAlike:    false,
			Compatibility: "compatible",
			Permissions:   []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:    []string{"License and copyright notice"},
			Limitations:   []string{"No liability", "No warranty"},
		},
		"APACHE-2.0": {
			License:       "Apache-2.0",
			LicenseType:   "permissive",
			IsOSIApproved: true,
			CommercialUse: true,
			Modification:  true,
			Distribution:  true,
			PatentGrant:   true,
			Attribution:   true,
			ShareAlike:    false,
			Compatibility: "compatible",
			Permissions:   []string{"Commercial use", "Modification", "Distribution", "Patent use", "Private use"},
			Conditions:    []string{"License and copyright notice", "State changes"},
			Limitations:   []string{"No liability", "No warranty", "No trademark use"},
		},
		"BSD-2-CLAUSE": {
			License:       "BSD-2-Clause",
			LicenseType:   "permissive",
			IsOSIApproved: true,
			CommercialUse: true,
			Modification:  true,
			Distribution:  true,
			PatentGrant:   false,
			Attribution:   true,
			ShareAlike:    false,
			Compatibility: "compatible",
			Permissions:   []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:    []string{"License and copyright notice"},
			Limitations:   []string{"No liability", "No warranty"},
		},
		"BSD-3-CLAUSE": {
			License:       "BSD-3-Clause",
			LicenseType:   "permissive",
			IsOSIApproved: true,
			CommercialUse: true,
			Modification:  true,
			Distribution:  true,
			PatentGrant:   false,
			Attribution:   true,
			ShareAlike:    false,
			Compatibility: "compatible",
			Permissions:   []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:    []string{"License and copyright notice"},
			Limitations:   []string{"No liability", "No warranty"},
		},
		"UNLICENSE": {
			License:       "Unlicense",
			LicenseType:   "permissive",
			IsOSIApproved: true,
			CommercialUse: true,
			Modification:  true,
			Distribution:  true,
			PatentGrant:   false,
			Attribution:   false,
			ShareAlike:    false,
			Compatibility: "compatible",
			Permissions:   []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:    []string{},
			Limitations:   []string{"No liability", "No warranty"},
		},
		"CC0-1.0": {
			License:       "CC0-1.0",
			LicenseType:   "permissive",
			IsOSIApproved: false,
			CommercialUse: true,
			Modification:  true,
			Distribution:  true,
			PatentGrant:   false,
			Attribution:   false,
			ShareAlike:    false,
			Compatibility: "compatible",
			Permissions:   []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:    []string{},
			Limitations:   []string{"No liability", "No warranty", "No patent rights"},
		},
		"WTFPL": {
			License:       "WTFPL",
			LicenseType:   "permissive",
			IsOSIApproved: false,
			CommercialUse: true,
			Modification:  true,
			Distribution:  true,
			PatentGrant:   false,
			Attribution:   false,
			ShareAlike:    false,
			Compatibility: "compatible",
			Permissions:   []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:    []string{},
			Limitations:   []string{},
		},
	}

	// Copyleft licenses - warning required
	copyleftLicenses := map[string]*LicenseInfo{
		"GPL-2.0": {
			License:          "GPL-2.0",
			LicenseType:      "copyleft",
			IsOSIApproved:    true,
			CommercialUse:    true,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       true,
			Compatibility:    "warning",
			CompatibilityMsg: "GPL-2.0 is a copyleft license. If you distribute this module, your entire application may need to be GPL-2.0 licensed.",
			Permissions:      []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:       []string{"Disclose source", "License and copyright notice", "Same license", "State changes"},
			Limitations:      []string{"No liability", "No warranty"},
		},
		"GPL-3.0": {
			License:          "GPL-3.0",
			LicenseType:      "copyleft",
			IsOSIApproved:    true,
			CommercialUse:    true,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      true,
			Attribution:      true,
			ShareAlike:       true,
			Compatibility:    "warning",
			CompatibilityMsg: "GPL-3.0 is a copyleft license. If you distribute this module, your entire application may need to be GPL-3.0 licensed.",
			Permissions:      []string{"Commercial use", "Modification", "Distribution", "Patent use", "Private use"},
			Conditions:       []string{"Disclose source", "License and copyright notice", "Same license", "State changes"},
			Limitations:      []string{"No liability", "No warranty"},
		},
		"LGPL-2.1": {
			License:          "LGPL-2.1",
			LicenseType:      "copyleft",
			IsOSIApproved:    true,
			CommercialUse:    true,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       true,
			Compatibility:    "warning",
			CompatibilityMsg: "LGPL-2.1 allows linking but modifications to the library must remain LGPL.",
			Permissions:      []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:       []string{"Disclose source (library)", "License and copyright notice", "Same license (library)"},
			Limitations:      []string{"No liability", "No warranty"},
		},
		"LGPL-3.0": {
			License:          "LGPL-3.0",
			LicenseType:      "copyleft",
			IsOSIApproved:    true,
			CommercialUse:    true,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      true,
			Attribution:      true,
			ShareAlike:       true,
			Compatibility:    "warning",
			CompatibilityMsg: "LGPL-3.0 allows linking but modifications to the library must remain LGPL.",
			Permissions:      []string{"Commercial use", "Modification", "Distribution", "Patent use", "Private use"},
			Conditions:       []string{"Disclose source (library)", "License and copyright notice", "Same license (library)"},
			Limitations:      []string{"No liability", "No warranty"},
		},
		"MPL-2.0": {
			License:          "MPL-2.0",
			LicenseType:      "copyleft",
			IsOSIApproved:    true,
			CommercialUse:    true,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      true,
			Attribution:      true,
			ShareAlike:       true,
			Compatibility:    "warning",
			CompatibilityMsg: "MPL-2.0 requires modified files to remain MPL, but allows combining with proprietary code.",
			Permissions:      []string{"Commercial use", "Modification", "Distribution", "Patent use", "Private use"},
			Conditions:       []string{"Disclose source (modified files)", "License and copyright notice", "Same license (modified files)"},
			Limitations:      []string{"No liability", "No warranty", "No trademark use"},
		},
	}

	// Network copyleft - strong warning
	networkCopyleftLicenses := map[string]*LicenseInfo{
		"AGPL-3.0": {
			License:          "AGPL-3.0",
			LicenseType:      "copyleft",
			IsOSIApproved:    true,
			CommercialUse:    true,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      true,
			Attribution:      true,
			ShareAlike:       true,
			Compatibility:    "warning",
			CompatibilityMsg: "AGPL-3.0 requires source disclosure even for network use. Your entire application may need to be AGPL licensed if users interact with it over a network.",
			Permissions:      []string{"Commercial use", "Modification", "Distribution", "Patent use", "Private use"},
			Conditions:       []string{"Disclose source", "License and copyright notice", "Network use is distribution", "Same license", "State changes"},
			Limitations:      []string{"No liability", "No warranty"},
		},
	}

	// Fair-code / Source-available licenses - special warning for n8n and similar
	fairCodeLicenses := map[string]*LicenseInfo{
		"SUSTAINABLE-USE": {
			License:          "Sustainable Use License",
			LicenseType:      "fair-code",
			IsOSIApproved:    false,
			CommercialUse:    false, // Restricted for SaaS
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       false,
			Compatibility:    "warning",
			CompatibilityMsg: "n8n Sustainable Use License: Free for internal use but CANNOT be offered as a hosted service to third parties without a commercial license from n8n.",
			Permissions:      []string{"Internal use", "Modification", "Distribution", "Private use"},
			Conditions:       []string{"Attribution", "No SaaS offering", "No competing service"},
			Limitations:      []string{"No commercial hosting", "No SaaS without license", "Not OSI approved"},
		},
		"N8N-SUSTAINABLE": {
			License:          "n8n Sustainable Use License",
			LicenseType:      "fair-code",
			IsOSIApproved:    false,
			CommercialUse:    false,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       false,
			Compatibility:    "warning",
			CompatibilityMsg: "n8n Sustainable Use License: Free for internal use but CANNOT be offered as a hosted service to third parties without a commercial license from n8n.",
			Permissions:      []string{"Internal use", "Modification", "Distribution", "Private use"},
			Conditions:       []string{"Attribution", "No SaaS offering", "No competing service"},
			Limitations:      []string{"No commercial hosting", "No SaaS without license", "Not OSI approved"},
		},
		"ELASTIC-2.0": {
			License:          "Elastic License 2.0",
			LicenseType:      "fair-code",
			IsOSIApproved:    false,
			CommercialUse:    true,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       false,
			Compatibility:    "warning",
			CompatibilityMsg: "Elastic License 2.0: Cannot provide the software as a managed service. Internal and commercial use allowed.",
			Permissions:      []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:       []string{"Attribution", "No managed service offering"},
			Limitations:      []string{"No managed service", "Not OSI approved"},
		},
		"SSPL-1.0": {
			License:          "Server Side Public License 1.0",
			LicenseType:      "fair-code",
			IsOSIApproved:    false,
			CommercialUse:    true,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       true,
			Compatibility:    "warning",
			CompatibilityMsg: "SSPL: If you offer the software as a service, you must release entire service stack as SSPL. Strongly restrictive for cloud deployments.",
			Permissions:      []string{"Commercial use", "Modification", "Distribution", "Private use"},
			Conditions:       []string{"Disclose entire service stack if offering as service", "Same license"},
			Limitations:      []string{"Service provider restrictions", "Not OSI approved"},
		},
		"BSL-1.1": {
			License:          "Business Source License 1.1",
			LicenseType:      "fair-code",
			IsOSIApproved:    false,
			CommercialUse:    false,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       false,
			Compatibility:    "warning",
			CompatibilityMsg: "Business Source License: Production use may require a commercial license. Check specific terms for allowed uses.",
			Permissions:      []string{"Non-production use", "Modification", "Distribution"},
			Conditions:       []string{"Commercial license for production", "Time-limited restrictions"},
			Limitations:      []string{"Production use restricted", "Not OSI approved"},
		},
		"COMMONS-CLAUSE": {
			License:          "Commons Clause",
			LicenseType:      "fair-code",
			IsOSIApproved:    false,
			CommercialUse:    false,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       false,
			Compatibility:    "warning",
			CompatibilityMsg: "Commons Clause: Cannot sell the software or provide it as a paid service without permission.",
			Permissions:      []string{"Non-commercial use", "Modification", "Distribution"},
			Conditions:       []string{"No selling", "No paid service"},
			Limitations:      []string{"No commercial sale", "Not OSI approved"},
		},
		"POLYFORM-NONCOMMERCIAL": {
			License:          "PolyForm Noncommercial",
			LicenseType:      "fair-code",
			IsOSIApproved:    false,
			CommercialUse:    false,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       false,
			Compatibility:    "warning",
			CompatibilityMsg: "PolyForm Noncommercial: Only for noncommercial purposes. Commercial use requires separate license.",
			Permissions:      []string{"Noncommercial use", "Modification", "Distribution"},
			Conditions:       []string{"Noncommercial only", "Attribution"},
			Limitations:      []string{"No commercial use", "Not OSI approved"},
		},
		"POLYFORM-SMALL-BUSINESS": {
			License:          "PolyForm Small Business",
			LicenseType:      "fair-code",
			IsOSIApproved:    false,
			CommercialUse:    true,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       false,
			Compatibility:    "warning",
			CompatibilityMsg: "PolyForm Small Business: Free for small businesses (under $1M revenue). Larger organizations need commercial license.",
			Permissions:      []string{"Small business use", "Modification", "Distribution"},
			Conditions:       []string{"Revenue limit", "Attribution"},
			Limitations:      []string{"Enterprise use restricted", "Not OSI approved"},
		},
	}

	// Check for exact matches and common variations
	normalizedLicense := strings.ReplaceAll(upperLicense, " ", "-")
	normalizedLicense = strings.ReplaceAll(normalizedLicense, "_", "-")

	// Check fair-code licenses FIRST (n8n, Elastic, SSPL, etc.)
	for key, licInfo := range fairCodeLicenses {
		if normalizedLicense == key || strings.Contains(normalizedLicense, key) {
			licInfo.License = license
			return licInfo
		}
	}

	// Check for n8n specific license patterns
	if strings.Contains(normalizedLicense, "N8N") ||
		strings.Contains(normalizedLicense, "SUSTAINABLE") ||
		strings.Contains(normalizedLicense, "FAIR-CODE") {
		return &LicenseInfo{
			License:          license,
			LicenseType:      "fair-code",
			IsOSIApproved:    false,
			CommercialUse:    false,
			Modification:     true,
			Distribution:     true,
			PatentGrant:      false,
			Attribution:      true,
			ShareAlike:       false,
			Compatibility:    "warning",
			CompatibilityMsg: "Fair-code/Sustainable Use License: Free for internal use but commercial hosting/SaaS may require a separate license.",
			Permissions:      []string{"Internal use", "Modification", "Distribution"},
			Conditions:       []string{"No SaaS offering without license", "Attribution"},
			Limitations:      []string{"Commercial hosting restricted", "Not OSI approved"},
		}
	}

	// Check network copyleft (AGPL contains GPL, so must check first)
	for key, licInfo := range networkCopyleftLicenses {
		if normalizedLicense == key || strings.Contains(normalizedLicense, key) {
			licInfo.License = license
			return licInfo
		}
	}

	// Check permissive licenses
	for key, licInfo := range permissiveLicenses {
		if normalizedLicense == key || strings.Contains(normalizedLicense, key) {
			licInfo.License = license // Preserve original
			return licInfo
		}
	}

	// Check copyleft licenses
	for key, licInfo := range copyleftLicenses {
		if normalizedLicense == key || strings.Contains(normalizedLicense, key) {
			licInfo.License = license
			return licInfo
		}
	}

	// Check for common proprietary indicators
	proprietaryIndicators := []string{"PROPRIETARY", "COMMERCIAL", "ALL-RIGHTS-RESERVED", "PRIVATE", "CLOSED"}
	for _, indicator := range proprietaryIndicators {
		if strings.Contains(normalizedLicense, indicator) {
			return &LicenseInfo{
				License:          license,
				LicenseType:      "proprietary",
				IsOSIApproved:    false,
				CommercialUse:    false,
				Modification:     false,
				Distribution:     false,
				Compatibility:    "incompatible",
				CompatibilityMsg: "Proprietary license. Redistribution and modification may not be permitted without explicit permission.",
			}
		}
	}

	// Unknown license
	info.CompatibilityMsg = fmt.Sprintf("License '%s' is not recognized. Please review the license terms manually.", license)
	return info
}

// validateStructure validates module structure
func (v *Validator) validateStructure(info *parser.ModuleInfo, result *ValidationResult) {
	if len(info.Nodes) == 0 {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "NO_NODES",
			Message:  "Module must contain at least one node",
			Severity: "error",
		})
		result.Score -= 30
	}

	// Check total size
	totalSize, err := v.calculateTotalSize(info.SourcePath)
	if err == nil {
		if totalSize > v.maxTotalSize {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "SIZE_EXCEEDED",
				Message:  fmt.Sprintf("Module size %d exceeds maximum %d", totalSize, v.maxTotalSize),
				Severity: "error",
			})
			result.Score -= 20
		}
	}
}

// validateNode validates a single node
func (v *Validator) validateNode(nodeInfo *parser.NodeInfo, basePath string, result *ValidationResult) {
	if nodeInfo.Type == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:     "MISSING_NODE_TYPE",
			Message:  "Node type is required",
			File:     nodeInfo.SourceFile,
			Severity: "error",
		})
		result.Score -= 10
	}

	// Check source file exists
	if nodeInfo.SourceFile != "" {
		sourcePath := filepath.Join(basePath, nodeInfo.SourceFile)
		if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "MISSING_SOURCE",
				Message:  fmt.Sprintf("Source file not found: %s", nodeInfo.SourceFile),
				File:     nodeInfo.SourceFile,
				Severity: "error",
			})
			result.Score -= 15
		} else {
			// Check file size
			info, _ := os.Stat(sourcePath)
			if info != nil && info.Size() > v.maxFileSize {
				result.Errors = append(result.Errors, ValidationError{
					Code:     "FILE_TOO_LARGE",
					Message:  fmt.Sprintf("Source file exceeds maximum size: %s", nodeInfo.SourceFile),
					File:     nodeInfo.SourceFile,
					Severity: "error",
				})
				result.Score -= 10
			}
		}
	}

	// Validate node type name
	if nodeInfo.Type != "" && !isValidNodeType(nodeInfo.Type) {
		result.Warnings = append(result.Warnings, ValidationError{
			Code:     "INVALID_NODE_TYPE",
			Message:  fmt.Sprintf("Node type contains invalid characters: %s", nodeInfo.Type),
			File:     nodeInfo.SourceFile,
			Severity: "warning",
		})
		result.Score -= 5
	}
}

// scanForSecurityIssues scans source files for security issues
func (v *Validator) scanForSecurityIssues(basePath string, result *ValidationResult) {
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip test, docs, examples, and node_modules directories
		if info.IsDir() {
			dirName := strings.ToLower(info.Name())
			switch dirName {
			case "test", "tests", "spec", "specs", "__tests__",
				"docs", "doc", "documentation",
				"examples", "example", "samples",
				"node_modules", ".git", "coverage":
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".js" && ext != ".ts" && ext != ".mjs" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		content := string(data)
		relPath, _ := filepath.Rel(basePath, path)

		// Check for blocked patterns
		v.checkBlockedPatterns(content, relPath, result)

		// Check for dangerous APIs
		v.checkDangerousAPIs(content, relPath, result)

		return nil
	})
}

// checkBlockedPatterns checks for blocked code patterns
func (v *Validator) checkBlockedPatterns(content, file string, result *ValidationResult) {
	for _, pattern := range v.blockedPatterns {
		if pattern.MatchString(content) {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "BLOCKED_PATTERN",
				Message:  fmt.Sprintf("Blocked code pattern found: %s", pattern.String()),
				File:     file,
				Severity: "error",
			})
			result.Score -= 25
		}
	}
}

// checkDangerousAPIs checks for dangerous API usage
func (v *Validator) checkDangerousAPIs(content, file string, result *ValidationResult) {
	// Network access
	if !v.allowNetworkAccess {
		networkPatterns := []string{
			`require\s*\(\s*['"]http['"]`,
			`require\s*\(\s*['"]https['"]`,
			`require\s*\(\s*['"]net['"]`,
			`require\s*\(\s*['"]dgram['"]`,
			`fetch\s*\(`,
			`XMLHttpRequest`,
			`WebSocket`,
		}
		for _, pattern := range networkPatterns {
			if matched, _ := regexp.MatchString(pattern, content); matched {
				result.Warnings = append(result.Warnings, ValidationError{
					Code:     "NETWORK_ACCESS",
					Message:  "Module may attempt network access",
					File:     file,
					Severity: "warning",
				})
				result.Score -= 10
				break
			}
		}
	}

	// File system access
	if !v.allowFileSystem {
		fsPatterns := []string{
			`require\s*\(\s*['"]fs['"]`,
			`require\s*\(\s*['"]fs/promises['"]`,
			`require\s*\(\s*['"]path['"]`,
		}
		for _, pattern := range fsPatterns {
			if matched, _ := regexp.MatchString(pattern, content); matched {
				result.Warnings = append(result.Warnings, ValidationError{
					Code:     "FILESYSTEM_ACCESS",
					Message:  "Module may attempt file system access",
					File:     file,
					Severity: "warning",
				})
				result.Score -= 10
				break
			}
		}
	}

	// Child process
	if !v.allowChildProcess {
		processPatterns := []string{
			`require\s*\(\s*['"]child_process['"]`,
			`exec\s*\(`,
			`spawn\s*\(`,
			`execSync\s*\(`,
			`spawnSync\s*\(`,
		}
		for _, pattern := range processPatterns {
			if matched, _ := regexp.MatchString(pattern, content); matched {
				result.Errors = append(result.Errors, ValidationError{
					Code:     "CHILD_PROCESS",
					Message:  "Module attempts to spawn child processes",
					File:     file,
					Severity: "error",
				})
				result.Score -= 30
				break
			}
		}
	}

	// Eval and dynamic code execution
	evalPatterns := []string{
		`\beval\s*\(`,
		`new\s+Function\s*\(`,
		`setTimeout\s*\(\s*['"]`,
		`setInterval\s*\(\s*['"]`,
	}
	for _, pattern := range evalPatterns {
		if matched, _ := regexp.MatchString(pattern, content); matched {
			result.Warnings = append(result.Warnings, ValidationError{
				Code:     "DYNAMIC_CODE",
				Message:  "Module uses dynamic code execution",
				File:     file,
				Severity: "warning",
			})
			result.Score -= 15
			break
		}
	}
}

// calculateTotalSize calculates total size of module files
func (v *Validator) calculateTotalSize(basePath string) (int64, error) {
	var total int64
	err := filepath.Walk(basePath, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

// compileBlockedPatterns compiles blocked code patterns
func compileBlockedPatterns() []*regexp.Regexp {
	patterns := []string{
		// Obfuscated code
		`\\x[0-9a-fA-F]{2}`,
		`\\u[0-9a-fA-F]{4}`,
		// Base64 encoded strings (potential payload)
		`atob\s*\(`,
		`btoa\s*\(`,
		// Process/global manipulation
		`process\.env`,
		`global\.__`,
		// Prototype pollution
		`__proto__`,
		`constructor\.prototype`,
	}

	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		if r, err := regexp.Compile(p); err == nil {
			compiled = append(compiled, r)
		}
	}
	return compiled
}

// isValidModuleName checks if module name is valid
func isValidModuleName(name string) bool {
	// npm-like naming: lowercase, dashes, no spaces
	pattern := regexp.MustCompile(`^[a-z0-9]([a-z0-9._-]*[a-z0-9])?$`)
	return pattern.MatchString(name) || strings.HasPrefix(name, "@")
}

// isValidNodeType checks if node type is valid
func isValidNodeType(nodeType string) bool {
	// Allow alphanumeric, dashes, underscores
	pattern := regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)
	return pattern.MatchString(nodeType)
}
