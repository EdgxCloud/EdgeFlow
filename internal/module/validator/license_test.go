package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeLicense(t *testing.T) {
	tests := []struct {
		name           string
		license        string
		expectedType   string
		expectedCompat string
		shareAlike     bool
		commercial     bool
	}{
		// Permissive licenses - compatible
		{"MIT", "MIT", "permissive", "compatible", false, true},
		{"ISC", "ISC", "permissive", "compatible", false, true},
		{"Apache-2.0", "Apache-2.0", "permissive", "compatible", false, true},
		{"BSD-3-Clause", "BSD-3-Clause", "permissive", "compatible", false, true},
		{"BSD-2-Clause", "BSD-2-Clause", "permissive", "compatible", false, true},
		{"Unlicense", "Unlicense", "permissive", "compatible", false, true},
		{"CC0-1.0", "CC0-1.0", "permissive", "compatible", false, true},
		{"WTFPL", "WTFPL", "permissive", "compatible", false, true},

		// Copyleft licenses - warning
		{"GPL-2.0", "GPL-2.0", "copyleft", "warning", true, true},
		{"GPL-3.0", "GPL-3.0", "copyleft", "warning", true, true},
		{"LGPL-2.1", "LGPL-2.1", "copyleft", "warning", true, true},
		{"LGPL-3.0", "LGPL-3.0", "copyleft", "warning", true, true},
		{"MPL-2.0", "MPL-2.0", "copyleft", "warning", true, true},
		{"AGPL-3.0", "AGPL-3.0", "copyleft", "warning", true, true},

		// Proprietary - incompatible
		{"Proprietary", "PROPRIETARY", "proprietary", "incompatible", false, false},
		{"All Rights Reserved", "All-Rights-Reserved", "proprietary", "incompatible", false, false},

		// Unknown
		{"Unknown", "Some-Unknown-License", "unknown", "unknown", false, false},
		{"Empty", "", "unspecified", "unknown", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := AnalyzeLicense(tt.license)

			assert.Equal(t, tt.expectedType, info.LicenseType, "License type mismatch")
			assert.Equal(t, tt.expectedCompat, info.Compatibility, "Compatibility mismatch")
			assert.Equal(t, tt.shareAlike, info.ShareAlike, "ShareAlike mismatch")
			assert.Equal(t, tt.commercial, info.CommercialUse, "CommercialUse mismatch")

			t.Logf("License: %s -> Type: %s, Compat: %s, Msg: %s",
				tt.license, info.LicenseType, info.Compatibility, info.CompatibilityMsg)
		})
	}
}

func TestAnalyzeLicenseVariations(t *testing.T) {
	// Test various case and format variations
	variations := []struct {
		input    string
		expected string
	}{
		{"mit", "permissive"},
		{"MIT", "permissive"},
		{"Mit", "permissive"},
		{"apache-2.0", "permissive"},
		{"Apache-2.0", "permissive"},
		{"APACHE-2.0", "permissive"},
		{"Apache 2.0", "permissive"},
		{"gpl-3.0", "copyleft"},
		{"GPL-3.0", "copyleft"},
		{"GPL 3.0", "copyleft"},
	}

	for _, tt := range variations {
		t.Run(tt.input, func(t *testing.T) {
			info := AnalyzeLicense(tt.input)
			assert.Equal(t, tt.expected, info.LicenseType,
				"License %s should be %s, got %s", tt.input, tt.expected, info.LicenseType)
		})
	}
}

func TestLicensePermissions(t *testing.T) {
	// Test that permissions are properly set
	mitInfo := AnalyzeLicense("MIT")
	assert.True(t, mitInfo.CommercialUse)
	assert.True(t, mitInfo.Modification)
	assert.True(t, mitInfo.Distribution)
	assert.True(t, mitInfo.Attribution)
	assert.False(t, mitInfo.ShareAlike)
	assert.Contains(t, mitInfo.Permissions, "Commercial use")
	assert.Contains(t, mitInfo.Conditions, "License and copyright notice")

	gplInfo := AnalyzeLicense("GPL-3.0")
	assert.True(t, gplInfo.CommercialUse)
	assert.True(t, gplInfo.Modification)
	assert.True(t, gplInfo.Distribution)
	assert.True(t, gplInfo.ShareAlike)
	assert.Contains(t, gplInfo.Conditions, "Disclose source")
	assert.Contains(t, gplInfo.Conditions, "Same license")

	apacheInfo := AnalyzeLicense("Apache-2.0")
	assert.True(t, apacheInfo.PatentGrant)
	assert.Contains(t, apacheInfo.Permissions, "Patent use")
}

func TestLicenseCompatibilityMessages(t *testing.T) {
	// Test that compatibility messages are helpful
	gplInfo := AnalyzeLicense("GPL-3.0")
	assert.Contains(t, gplInfo.CompatibilityMsg, "copyleft")
	assert.Contains(t, gplInfo.CompatibilityMsg, "GPL")

	agplInfo := AnalyzeLicense("AGPL-3.0")
	assert.Contains(t, agplInfo.CompatibilityMsg, "network")
	assert.Contains(t, agplInfo.CompatibilityMsg, "AGPL")

	propInfo := AnalyzeLicense("PROPRIETARY")
	assert.Contains(t, propInfo.CompatibilityMsg, "Proprietary")
	assert.Contains(t, propInfo.CompatibilityMsg, "permission")

	unknownInfo := AnalyzeLicense("CustomLicense-1.0")
	assert.Contains(t, unknownInfo.CompatibilityMsg, "not recognized")
}

func TestNodeREDCommonLicenses(t *testing.T) {
	// Test licenses commonly found in Node-RED modules
	commonLicenses := []string{
		"MIT",
		"Apache-2.0",
		"ISC",
		"BSD-3-Clause",
	}

	for _, license := range commonLicenses {
		t.Run(license, func(t *testing.T) {
			info := AnalyzeLicense(license)
			assert.Equal(t, "compatible", info.Compatibility,
				"Common Node-RED license %s should be compatible", license)
			assert.True(t, info.CommercialUse,
				"Common Node-RED license %s should allow commercial use", license)
		})
	}
}
