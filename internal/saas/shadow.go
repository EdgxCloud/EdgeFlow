package saas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

// ShadowManager manages device shadow state synchronization
type ShadowManager struct {
	config        *Config
	logger        *zap.Logger
	currentShadow *Shadow
	mu            sync.RWMutex

	// Callbacks for desired state changes
	onDesiredChange func(delta map[string]interface{})
}

// NewShadowManager creates a shadow manager
func NewShadowManager(config *Config, logger *zap.Logger) *ShadowManager {
	return &ShadowManager{
		config: config,
		logger: logger,
	}
}

// SetDesiredChangeHandler sets callback for when cloud updates desired state
func (s *ShadowManager) SetDesiredChangeHandler(handler func(delta map[string]interface{})) {
	s.onDesiredChange = handler
}

// GetShadow retrieves the current device shadow from SaaS
func (s *ShadowManager) GetShadow() (*Shadow, error) {
	if !s.config.IsProvisioned() {
		return nil, ErrNotProvisioned("cannot get shadow")
	}

	url := fmt.Sprintf("%s/devices/%s/shadow", s.config.APIURL(), s.config.DeviceID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-API-Key", s.config.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get shadow failed: %d", resp.StatusCode)
	}

	var shadow Shadow
	if err := json.NewDecoder(resp.Body).Decode(&shadow); err != nil {
		return nil, err
	}

	s.mu.Lock()
	s.currentShadow = &shadow
	s.mu.Unlock()

	return &shadow, nil
}

// UpdateReported updates the reported state (device → cloud)
func (s *ShadowManager) UpdateReported(state map[string]interface{}) error {
	if !s.config.IsProvisioned() {
		return ErrNotProvisioned("cannot update shadow")
	}

	url := fmt.Sprintf("%s/devices/%s/shadow", s.config.APIURL(), s.config.DeviceID)

	payload := map[string]interface{}{
		"reported": state,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", s.config.APIKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update shadow failed: %d", resp.StatusCode)
	}

	// Update local cache
	var shadow Shadow
	if err := json.NewDecoder(resp.Body).Decode(&shadow); err == nil {
		s.mu.Lock()
		s.currentShadow = &shadow
		s.mu.Unlock()

		// Notify if there's a new delta
		if len(shadow.Delta) > 0 && s.onDesiredChange != nil {
			s.onDesiredChange(shadow.Delta)
		}
	}

	return nil
}

// GetCurrentShadow returns the cached shadow (no network call)
func (s *ShadowManager) GetCurrentShadow() *Shadow {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentShadow
}

// GetDelta returns the current delta (desired - reported)
func (s *ShadowManager) GetDelta() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.currentShadow == nil {
		return nil
	}

	return s.currentShadow.Delta
}

// StartPeriodicSync starts background sync every interval
func (s *ShadowManager) StartPeriodicSync(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			if _, err := s.GetShadow(); err != nil {
				s.logger.Error("Shadow sync failed", zap.Error(err))
			} else {
				s.logger.Debug("Shadow synced")
			}
		}
	}()
}
