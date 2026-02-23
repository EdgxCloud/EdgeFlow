package network

import (
	"context"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// TCP Client Tests
// ============================================================================

func TestNewTCPClientExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"host": "localhost",
				"port": 8080,
			},
			wantErr: false,
		},
		{
			name: "missing host",
			config: map[string]interface{}{
				"port": 8080,
			},
			wantErr: true,
			errMsg:  "host is required",
		},
		{
			name: "missing port",
			config: map[string]interface{}{
				"host": "localhost",
			},
			wantErr: true,
			errMsg:  "port is required",
		},
		{
			name: "full config with all options",
			config: map[string]interface{}{
				"host":           "192.168.1.100",
				"port":           9000,
				"autoReconnect":  true,
				"reconnectDelay": 3000,
				"timeout":        30,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewTCPClientExecutor(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestTCPClientConfig_Defaults(t *testing.T) {
	config := map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	}

	executor, err := NewTCPClientExecutor(config)
	require.NoError(t, err)

	tcpExecutor := executor.(*TCPClientExecutor)
	assert.Equal(t, 5000, tcpExecutor.config.ReconnectDelay)
	assert.Equal(t, 10, tcpExecutor.config.Timeout)
}

func TestTCPClient_ConnectAndSend_MockServer(t *testing.T) {
	// Start a mock TCP server
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	var wg sync.WaitGroup
	var receivedData string

	// Server goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		receivedData = string(buf[:n])
	}()

	// Create TCP client executor
	config := map[string]interface{}{
		"host":    "127.0.0.1",
		"port":    addr.Port,
		"timeout": 5,
	}

	executor, err := NewTCPClientExecutor(config)
	require.NoError(t, err)
	defer executor.Cleanup()

	// Send data
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := node.Message{
		Payload: map[string]interface{}{
			"send": "Hello TCP Server",
		},
	}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	payload, ok := result.Payload.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, payload["sent"].(bool))

	// Wait for server to receive data
	wg.Wait()

	assert.Contains(t, receivedData, "Hello TCP Server")
}

func TestTCPClient_ReceiveData_MockServer(t *testing.T) {
	// Start a mock TCP server that sends data
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	// Server goroutine - sends data after connection
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		time.Sleep(100 * time.Millisecond)
		conn.Write([]byte("Server Response\n"))
	}()

	// Create TCP client executor
	config := map[string]interface{}{
		"host":    "127.0.0.1",
		"port":    addr.Port,
		"timeout": 5,
	}

	executor, err := NewTCPClientExecutor(config)
	require.NoError(t, err)
	defer executor.Cleanup()

	// Execute without send data - should receive server data
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	msg := node.Message{
		Payload: map[string]interface{}{},
	}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	payload, ok := result.Payload.(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, payload["data"].(string), "Server Response")
}

func TestTCPClient_ConnectionFailure(t *testing.T) {
	// Try to connect to a non-existent server
	config := map[string]interface{}{
		"host":    "127.0.0.1",
		"port":    59999, // Unlikely to be in use
		"timeout": 1,
	}

	executor, err := NewTCPClientExecutor(config)
	require.NoError(t, err)
	defer executor.Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	msg := node.Message{
		Payload: map[string]interface{}{
			"send": "test",
		},
	}

	_, err = executor.Execute(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect")
}

func TestTCPClient_Cleanup(t *testing.T) {
	config := map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	}

	executor, err := NewTCPClientExecutor(config)
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}

// ============================================================================
// UDP Tests
// ============================================================================

func TestNewUDPExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid listen config",
			config: map[string]interface{}{
				"mode": "listen",
				"port": 8080,
			},
			wantErr: false,
		},
		{
			name: "valid send config",
			config: map[string]interface{}{
				"mode": "send",
				"host": "localhost",
				"port": 8080,
			},
			wantErr: false,
		},
		{
			name: "missing port",
			config: map[string]interface{}{
				"mode": "listen",
			},
			wantErr: true,
			errMsg:  "port is required",
		},
		{
			name: "invalid mode",
			config: map[string]interface{}{
				"mode": "invalid",
				"port": 8080,
			},
			wantErr: true,
			errMsg:  "mode must be 'listen' or 'send'",
		},
		{
			name: "full config with buffer size",
			config: map[string]interface{}{
				"mode":       "listen",
				"port":       8080,
				"bufferSize": 8192,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewUDPExecutor(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestUDPConfig_Defaults(t *testing.T) {
	config := map[string]interface{}{
		"port": 8080,
	}

	executor, err := NewUDPExecutor(config)
	require.NoError(t, err)

	udpExecutor := executor.(*UDPExecutor)
	assert.Equal(t, "listen", udpExecutor.config.Mode)
	assert.Equal(t, 4096, udpExecutor.config.BufferSize)
}

func TestUDP_ListenAndReceive_MockClient(t *testing.T) {
	// Create UDP listener executor
	config := map[string]interface{}{
		"mode": "listen",
		"port": 0, // Let OS assign port
	}

	executor, err := NewUDPExecutor(config)
	require.NoError(t, err)
	defer executor.Cleanup()

	udpExecutor := executor.(*UDPExecutor)

	// Setup the listener
	err = udpExecutor.setup()
	require.NoError(t, err)

	// Get assigned port
	addr := udpExecutor.conn.LocalAddr().(*net.UDPAddr)

	// Start read loop
	go udpExecutor.readLoop()

	// Send data from a client
	clientConn, err := net.Dial("udp", addr.String())
	require.NoError(t, err)
	defer clientConn.Close()

	_, err = clientConn.Write([]byte("Hello UDP Server"))
	require.NoError(t, err)

	// Wait for message in output channel
	select {
	case msg := <-udpExecutor.outputChan:
		payload, ok := msg.Payload.(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, payload["data"].(string), "Hello UDP Server")
		assert.NotEmpty(t, payload["from"])
		assert.Greater(t, payload["size"].(int), 0)
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for UDP message")
	}
}

func TestUDP_SendData_MockServer(t *testing.T) {
	// Start a mock UDP server
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	require.NoError(t, err)

	serverConn, err := net.ListenUDP("udp", serverAddr)
	require.NoError(t, err)
	defer serverConn.Close()

	addr := serverConn.LocalAddr().(*net.UDPAddr)

	var receivedData string
	var wg sync.WaitGroup
	wg.Add(1)

	// Server goroutine
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		serverConn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		receivedData = string(buf[:n])
	}()

	// Create UDP send executor
	config := map[string]interface{}{
		"mode": "send",
		"host": "127.0.0.1",
		"port": addr.Port,
	}

	executor, err := NewUDPExecutor(config)
	require.NoError(t, err)
	defer executor.Cleanup()

	// Send data
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	msg := node.Message{
		Payload: map[string]interface{}{
			"send": "Hello UDP Client",
		},
	}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	payload, ok := result.Payload.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, payload["sent"].(bool))

	// Wait for server to receive data
	wg.Wait()

	assert.Equal(t, "Hello UDP Client", receivedData)
}

func TestUDP_SendWithDynamicAddress(t *testing.T) {
	// Start a mock UDP server
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	require.NoError(t, err)

	serverConn, err := net.ListenUDP("udp", serverAddr)
	require.NoError(t, err)
	defer serverConn.Close()

	addr := serverConn.LocalAddr().(*net.UDPAddr)

	var receivedData string
	var wg sync.WaitGroup
	wg.Add(1)

	// Server goroutine
	go func() {
		defer wg.Done()
		buf := make([]byte, 1024)
		serverConn.SetReadDeadline(time.Now().Add(3 * time.Second))
		n, _, err := serverConn.ReadFromUDP(buf)
		if err != nil {
			return
		}
		receivedData = string(buf[:n])
	}()

	// Create UDP send executor with a default address
	config := map[string]interface{}{
		"mode": "send",
		"host": "127.0.0.1",
		"port": addr.Port,
	}

	executor, err := NewUDPExecutor(config)
	require.NoError(t, err)
	defer executor.Cleanup()

	// Send data with override address in message
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	msg := node.Message{
		Payload: map[string]interface{}{
			"send": "Dynamic Address Message",
			"host": "127.0.0.1",
			"port": float64(addr.Port), // JSON numbers are float64
		},
	}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	payload, ok := result.Payload.(map[string]interface{})
	require.True(t, ok)
	assert.True(t, payload["sent"].(bool))

	// Wait for server to receive data
	wg.Wait()

	assert.Equal(t, "Dynamic Address Message", receivedData)
}

func TestUDP_SendNoData_Error(t *testing.T) {
	config := map[string]interface{}{
		"mode": "send",
		"host": "127.0.0.1",
		"port": 8080,
	}

	executor, err := NewUDPExecutor(config)
	require.NoError(t, err)
	defer executor.Cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Message without "send" key
	msg := node.Message{
		Payload: map[string]interface{}{},
	}

	_, err = executor.Execute(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no data to send")
}

func TestUDP_ListenMode_ContextCancellation(t *testing.T) {
	config := map[string]interface{}{
		"mode": "listen",
		"port": 0,
	}

	executor, err := NewUDPExecutor(config)
	require.NoError(t, err)
	defer executor.Cleanup()

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	msg := node.Message{
		Payload: map[string]interface{}{},
	}

	_, err = executor.Execute(ctx, msg)
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestUDP_Cleanup(t *testing.T) {
	config := map[string]interface{}{
		"mode": "listen",
		"port": 8080,
	}

	executor, err := NewUDPExecutor(config)
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}

func TestUDP_SendMode_MissingHost(t *testing.T) {
	config := map[string]interface{}{
		"mode": "send",
		"port": 8080,
		// No host specified
	}

	executor, err := NewUDPExecutor(config)
	require.NoError(t, err)
	defer executor.Cleanup()

	udpExecutor := executor.(*UDPExecutor)
	err = udpExecutor.setup()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host is required for send mode")
}

// ============================================================================
// Benchmarks
// ============================================================================

func BenchmarkTCPClientExecutor_Create(b *testing.B) {
	config := map[string]interface{}{
		"host": "localhost",
		"port": 8080,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor, _ := NewTCPClientExecutor(config)
		if executor != nil {
			executor.Cleanup()
		}
	}
}

func BenchmarkUDPExecutor_Create(b *testing.B) {
	config := map[string]interface{}{
		"mode": "listen",
		"port": 8080,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor, _ := NewUDPExecutor(config)
		if executor != nil {
			executor.Cleanup()
		}
	}
}
