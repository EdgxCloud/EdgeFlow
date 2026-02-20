package saas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
)

// ProvisioningClient handles device provisioning
type ProvisioningClient struct {
	config *Config
	logger *zap.Logger
}

// NewProvisioningClient creates a provisioning client
func NewProvisioningClient(config *Config, logger *zap.Logger) *ProvisioningClient {
	return &ProvisioningClient{
		config: config,
		logger: logger,
	}
}

// Provision registers this device with the SaaS platform
func (p *ProvisioningClient) Provision() (*ProvisionResponse, error) {
	if p.config.ProvisioningCode == "" {
		return nil, ErrInvalidConfig("provisioning_code is required")
	}

	p.logger.Info("Starting device provisioning",
		zap.String("server", p.config.ServerURL))

	// Gather device information
	hwInfo := p.gatherHardwareInfo()
	netInfo := p.gatherNetworkInfo()

	req := &ProvisionRequest{
		ProvisioningCode: p.config.ProvisioningCode,
		HardwareInfo:     hwInfo,
		NetworkInfo:      netInfo,
	}

	// Send provisioning request
	url := p.config.APIURL() + "/devices/provision"
	resp, err := p.postJSON(url, req)
	if err != nil {
		return nil, ErrProvisioningFailed(err)
	}

	var provResp ProvisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&provResp); err != nil {
		return nil, ErrProvisioningFailed(err)
	}

	p.logger.Info("Device provisioned successfully",
		zap.String("device_id", provResp.DeviceID))

	return &provResp, nil
}

// gatherHardwareInfo collects system hardware information
func (p *ProvisioningClient) gatherHardwareInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// OS and architecture
	info["os"] = runtime.GOOS
	info["arch"] = runtime.GOARCH
	info["kernel"] = "" // Can add syscall.Uname() on Linux

	// Try to read board model (Raspberry Pi specific)
	if boardModel, err := os.ReadFile("/proc/device-tree/model"); err == nil {
		info["board_model"] = strings.TrimSpace(string(boardModel))
	} else {
		info["board_model"] = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// CPU info
	info["cpu_cores"] = runtime.NumCPU()

	// Memory (simplified, would need syscall for exact values)
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	info["ram_mb"] = m.Sys / 1024 / 1024

	return info
}

// gatherNetworkInfo collects network configuration
func (p *ProvisioningClient) gatherNetworkInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// Hostname
	if hostname, err := os.Hostname(); err == nil {
		info["hostname"] = hostname
	}

	// Get primary network interface
	ifaces, err := net.Interfaces()
	if err == nil {
		for _, iface := range ifaces {
			// Skip loopback and down interfaces
			if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
				continue
			}

			// Get addresses
			addrs, err := iface.Addrs()
			if err != nil || len(addrs) == 0 {
				continue
			}

			// Use first valid interface
			info["mac_address"] = iface.HardwareAddr.String()

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
					if ipnet.IP.To4() != nil {
						info["ip_address"] = ipnet.IP.String()
						break
					}
				}
			}

			if _, ok := info["ip_address"]; ok {
				break
			}
		}
	}

	// Connection type detection (simplified)
	if strings.Contains(fmt.Sprint(info["mac_address"]), "wlan") {
		info["connection_type"] = "wifi"
	} else {
		info["connection_type"] = "ethernet"
	}

	return info
}

// postJSON sends a POST request with JSON body
func (p *ProvisioningClient) postJSON(url string, body interface{}) (*http.Response, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * http.DefaultClient.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("provisioning failed: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}
