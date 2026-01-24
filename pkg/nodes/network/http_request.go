package network

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// HTTPRequestConfig نود HTTP Request
type HTTPRequestConfig struct {
	Method         string            `json:"method"`         // GET, POST, PUT, DELETE, PATCH
	URL            string            `json:"url"`            // URL
	Headers        map[string]string `json:"headers"`        // Headers
	Timeout        int               `json:"timeout"`        // Timeout in seconds
	RetryCount     int               `json:"retryCount"`     // Number of retries
	RetryDelay     int               `json:"retryDelay"`     // Delay between retries (ms)
	FollowRedirect bool              `json:"followRedirect"` // Follow redirects

	// Cookie handling
	UseCookies bool `json:"useCookies"` // Enable cookie jar

	// Proxy configuration
	ProxyURL      string `json:"proxyUrl"`      // Proxy server URL
	ProxyUsername string `json:"proxyUsername"` // Proxy authentication username
	ProxyPassword string `json:"proxyPassword"` // Proxy authentication password

	// TLS/SSL configuration
	CACertPath     string `json:"caCertPath"`     // Custom CA certificate path
	InsecureSkipTLS bool  `json:"insecureSkipTls"` // Skip TLS verification (not recommended)

	// Basic Auth
	BasicAuthUsername string `json:"basicAuthUsername"` // Basic auth username
	BasicAuthPassword string `json:"basicAuthPassword"` // Basic auth password

	// OAuth2 configuration
	OAuth2        *OAuth2Config `json:"oauth2"`        // OAuth2 configuration
}

// OAuth2Config OAuth2 configuration for HTTP requests
type OAuth2Config struct {
	TokenURL     string            `json:"tokenUrl"`     // Token endpoint URL
	ClientID     string            `json:"clientId"`     // Client ID
	ClientSecret string            `json:"clientSecret"` // Client secret
	Scope        string            `json:"scope"`        // Requested scope
	GrantType    string            `json:"grantType"`    // Grant type (client_credentials, password, refresh_token)
	Username     string            `json:"username"`     // Resource owner username (password grant)
	Password     string            `json:"password"`     // Resource owner password (password grant)
	RefreshToken string            `json:"refreshToken"` // Refresh token
	ExtraParams  map[string]string `json:"extraParams"`  // Extra token request parameters
}

// HTTPRequestExecutor اجراکننده نود HTTP Request
type HTTPRequestExecutor struct {
	config      HTTPRequestConfig
	client      *http.Client
	cookieJar   http.CookieJar
	accessToken string
	tokenExpiry time.Time
	tokenMu     sync.RWMutex
}

// NewHTTPRequestExecutor ایجاد HTTPRequestExecutor
func NewHTTPRequestExecutor() node.Executor {
	return &HTTPRequestExecutor{}
}

// Init initializes the executor with configuration
func (e *HTTPRequestExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var httpConfig HTTPRequestConfig
	if err := json.Unmarshal(configJSON, &httpConfig); err != nil {
		return fmt.Errorf("invalid http config: %w", err)
	}

	// Default values
	if httpConfig.Method == "" {
		httpConfig.Method = "GET"
	}
	if httpConfig.Timeout == 0 {
		httpConfig.Timeout = 30
	}
	if httpConfig.RetryCount < 0 {
		httpConfig.RetryCount = 0
	}
	if httpConfig.RetryDelay == 0 {
		httpConfig.RetryDelay = 1000
	}

	// Create transport with proxy and TLS config
	transport := &http.Transport{}

	// Configure proxy
	if httpConfig.ProxyURL != "" {
		proxyURL, err := url.Parse(httpConfig.ProxyURL)
		if err != nil {
			return fmt.Errorf("invalid proxy URL: %w", err)
		}

		// Add proxy authentication if provided
		if httpConfig.ProxyUsername != "" {
			proxyURL.User = url.UserPassword(httpConfig.ProxyUsername, httpConfig.ProxyPassword)
		}

		transport.Proxy = http.ProxyURL(proxyURL)
	}

	// Configure TLS
	tlsConfig := &tls.Config{}
	if httpConfig.InsecureSkipTLS {
		tlsConfig.InsecureSkipVerify = true
	}

	// Load custom CA certificate
	if httpConfig.CACertPath != "" {
		caCert, err := os.ReadFile(httpConfig.CACertPath)
		if err != nil {
			return fmt.Errorf("failed to read CA certificate: %w", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	transport.TLSClientConfig = tlsConfig

	// Create cookie jar if enabled
	var jar http.CookieJar
	if httpConfig.UseCookies {
		jar, _ = cookiejar.New(nil)
	}

	// Create HTTP client
	client := &http.Client{
		Timeout:   time.Duration(httpConfig.Timeout) * time.Second,
		Transport: transport,
		Jar:       jar,
	}

	if !httpConfig.FollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	e.config = httpConfig
	e.client = client
	e.cookieJar = jar
	return nil
}

// Execute اجرای نود
func (e *HTTPRequestExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get URL from config or message
	url := e.config.URL
	payload := msg.Payload
	if urlFromMsg, ok := payload["url"].(string); ok && urlFromMsg != "" {
		url = urlFromMsg
	}

	if url == "" {
		return node.Message{}, fmt.Errorf("URL is required")
	}

	// Get method from config or message
	method := e.config.Method
	if methodFromMsg, ok := payload["method"].(string); ok && methodFromMsg != "" {
		method = strings.ToUpper(methodFromMsg)
	}

	// Get body from message
	var body io.Reader
	if bodyData, ok := payload["body"]; ok {
		switch v := bodyData.(type) {
		case string:
			body = strings.NewReader(v)
		case map[string]interface{}, []interface{}:
			jsonData, err := json.Marshal(v)
			if err != nil {
				return node.Message{}, fmt.Errorf("failed to marshal body: %w", err)
			}
			body = bytes.NewReader(jsonData)
		}
	}

	// Execute with retry
	var lastErr error
	for attempt := 0; attempt <= e.config.RetryCount; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(e.config.RetryDelay) * time.Millisecond)
		}

		resp, err := e.executeRequest(ctx, method, url, body)
		if err == nil {
			return resp, nil
		}

		lastErr = err
	}

	return node.Message{}, fmt.Errorf("request failed after %d retries: %w", e.config.RetryCount, lastErr)
}

// executeRequest اجرای یک درخواست HTTP
func (e *HTTPRequestExecutor) executeRequest(ctx context.Context, method, url string, body io.Reader) (node.Message, error) {
	// Create request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range e.config.Headers {
		req.Header.Set(key, value)
	}

	// Set default Content-Type for POST/PUT/PATCH
	if (method == "POST" || method == "PUT" || method == "PATCH") && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add Basic Auth if configured
	if e.config.BasicAuthUsername != "" {
		auth := base64.StdEncoding.EncodeToString([]byte(e.config.BasicAuthUsername + ":" + e.config.BasicAuthPassword))
		req.Header.Set("Authorization", "Basic "+auth)
	}

	// Add OAuth2 token if configured
	if e.config.OAuth2 != nil {
		token, err := e.getOAuth2Token(ctx)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to get OAuth2 token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Execute request
	startTime := time.Now()
	resp, err := e.client.Do(req)
	if err != nil {
		return node.Message{}, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(startTime)

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return node.Message{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response body as JSON if possible
	var parsedBody interface{}
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			// If JSON parsing fails, use raw string
			parsedBody = string(respBody)
		}
	} else {
		parsedBody = string(respBody)
	}

	// Create response message
	responseMsg := node.Message{
		Payload: map[string]interface{}{
			"statusCode": resp.StatusCode,
			"headers":    resp.Header,
			"body":       parsedBody,
			"duration":   duration.Milliseconds(),
		},
	}

	return responseMsg, nil
}

// Cleanup پاکسازی منابع
func (e *HTTPRequestExecutor) Cleanup() error {
	if e.client != nil {
		e.client.CloseIdleConnections()
	}
	return nil
}

// getOAuth2Token retrieves or refreshes OAuth2 access token
func (e *HTTPRequestExecutor) getOAuth2Token(ctx context.Context) (string, error) {
	e.tokenMu.RLock()
	// Check if we have a valid token
	if e.accessToken != "" && time.Now().Before(e.tokenExpiry) {
		token := e.accessToken
		e.tokenMu.RUnlock()
		return token, nil
	}
	e.tokenMu.RUnlock()

	// Need to get a new token
	e.tokenMu.Lock()
	defer e.tokenMu.Unlock()

	// Double-check after acquiring write lock
	if e.accessToken != "" && time.Now().Before(e.tokenExpiry) {
		return e.accessToken, nil
	}

	oauth2Config := e.config.OAuth2
	if oauth2Config == nil || oauth2Config.TokenURL == "" {
		return "", fmt.Errorf("OAuth2 token URL is required")
	}

	// Build token request body
	data := url.Values{}
	data.Set("client_id", oauth2Config.ClientID)
	data.Set("client_secret", oauth2Config.ClientSecret)

	if oauth2Config.Scope != "" {
		data.Set("scope", oauth2Config.Scope)
	}

	// Set grant type and additional parameters
	grantType := oauth2Config.GrantType
	if grantType == "" {
		grantType = "client_credentials"
	}
	data.Set("grant_type", grantType)

	switch grantType {
	case "password":
		data.Set("username", oauth2Config.Username)
		data.Set("password", oauth2Config.Password)
	case "refresh_token":
		data.Set("refresh_token", oauth2Config.RefreshToken)
	}

	// Add extra parameters
	for key, value := range oauth2Config.ExtraParams {
		data.Set(key, value)
	}

	// Create token request
	req, err := http.NewRequestWithContext(ctx, "POST", oauth2Config.TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Execute token request with a simple client (no proxy/auth)
	tokenClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := tokenClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse token response
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	// Store token with expiry (subtract 60 seconds for safety margin)
	e.accessToken = tokenResp.AccessToken
	if tokenResp.ExpiresIn > 0 {
		e.tokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn-60) * time.Second)
	} else {
		// Default to 1 hour if not specified
		e.tokenExpiry = time.Now().Add(1 * time.Hour)
	}

	return e.accessToken, nil
}

// GetCookies returns cookies for a given URL (useful for debugging)
func (e *HTTPRequestExecutor) GetCookies(urlStr string) []*http.Cookie {
	if e.cookieJar == nil {
		return nil
	}
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil
	}
	return e.cookieJar.Cookies(parsedURL)
}
