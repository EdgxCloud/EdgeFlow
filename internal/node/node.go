package node

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// MessageType defines the type of message being passed between nodes
type MessageType string

const (
	MessageTypeData  MessageType = "data"
	MessageTypeError MessageType = "error"
	MessageTypeEvent MessageType = "event"
)

// Message represents data flowing between nodes
type Message struct {
	Type    MessageType            `json:"type"`
	Payload map[string]interface{} `json:"payload"`
	Topic   string                 `json:"topic,omitempty"`
	Error   error                  `json:"error,omitempty"`

	// Ports selects which of the node's output ports this message is routed to.
	// Routing nodes (switch, if) set it to the ports their rules matched.
	//
	//	nil            -> broadcast to every connected output (default)
	//	[]int{}        -> drop; the message matched no output
	//	[]int{0, 2}    -> deliver only to links on output ports 0 and 2
	//
	// nil is the default so that ordinary single-output nodes keep broadcasting
	// without having to opt in.
	Ports []int `json:"-"`
}

// RouteTo returns a copy of the message routed to the given output ports.
// Passing no ports drops the message.
func (m Message) RouteTo(ports ...int) Message {
	if ports == nil {
		ports = []int{}
	}
	m.Ports = ports
	return m
}

// NodeType defines the category of a node
type NodeType string

const (
	NodeTypeInput      NodeType = "input"
	NodeTypeOutput     NodeType = "output"
	NodeTypeProcessing NodeType = "processing"
	NodeTypeFunction   NodeType = "function"
)

// NodeStatus represents the current state of a node
type NodeStatus string

const (
	NodeStatusIdle    NodeStatus = "idle"
	NodeStatusRunning NodeStatus = "running"
	NodeStatusError   NodeStatus = "error"
)

// ExecutionEvent represents a single node execution result for debugging
type ExecutionEvent struct {
	NodeID        string                 `json:"node_id"`
	NodeName      string                 `json:"node_name"`
	NodeType      string                 `json:"node_type"`
	Input         map[string]interface{} `json:"input"`
	Output        map[string]interface{} `json:"output"`
	Status        string                 `json:"status"` // "success" or "error"
	Error         string                 `json:"error,omitempty"`
	ExecutionTime int64                  `json:"execution_time"` // milliseconds
	Timestamp     int64                  `json:"timestamp"`
}

// ExecutionCallback is called after each node execution with the result
type ExecutionCallback func(event ExecutionEvent)

// Node represents a single processing unit in a flow
type Node struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Category    NodeType               `json:"category"`
	Config      map[string]interface{} `json:"config"`
	Inputs      []string               `json:"inputs"`
	Outputs     []string               `json:"outputs"`
	Status      NodeStatus             `json:"status"`
	mu          sync.RWMutex
	executor    Executor
	inputChan   chan Message
	outputChans []outputLink
	ctx         context.Context
	cancel      context.CancelFunc
	onExecution ExecutionCallback
}

// outputLink is a single outgoing edge: the target's input channel plus the
// output port on this node that the edge leaves from.
type outputLink struct {
	ch   chan Message
	port int
}

// Executor defines the interface for node execution logic
type Executor interface {
	Execute(ctx context.Context, msg Message) (Message, error)
	Init(config map[string]interface{}) error
	Cleanup() error
}

// SelfTriggering is an optional interface for executors that generate their own messages
// (e.g., inject/timer nodes). The Run method is called in a goroutine after Init.
// It should send messages by calling the provided send function.
type SelfTriggering interface {
	Run(ctx context.Context, send func(Message))
}

// NewNode creates a new node instance
func NewNode(nodeType, name string, category NodeType, executor Executor) *Node {
	return &Node{
		ID:          uuid.New().String(),
		Type:        nodeType,
		Name:        name,
		Category:    category,
		Config:      make(map[string]interface{}),
		Inputs:      []string{},
		Outputs:     []string{},
		Status:      NodeStatusIdle,
		executor:    executor,
		inputChan:   make(chan Message, 100),
		outputChans: []outputLink{},
	}
}

// Start begins processing messages
func (n *Node) Start(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Status == NodeStatusRunning {
		return fmt.Errorf("node %s is already running", n.ID)
	}

	n.ctx, n.cancel = context.WithCancel(ctx)
	n.Status = NodeStatusRunning

	// Initialize executor
	if err := n.executor.Init(n.Config); err != nil {
		n.Status = NodeStatusError
		return fmt.Errorf("failed to initialize node: %w", err)
	}

	// Start message processing goroutine
	go n.process()

	// If executor is self-triggering (e.g., inject/timer), start its Run loop
	if st, ok := n.executor.(SelfTriggering); ok {
		go st.Run(n.ctx, func(msg Message) {
			n.handleMessage(msg)
		})
	}

	return nil
}

// Stop halts message processing
func (n *Node) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Status != NodeStatusRunning {
		return nil
	}

	if n.cancel != nil {
		n.cancel()
	}

	n.Status = NodeStatusIdle

	// Cleanup executor
	return n.executor.Cleanup()
}

// Send sends a message to this node
func (n *Node) Send(msg Message) error {
	// ctx is only set by Start; every case of a select is evaluated, so reading
	// n.ctx.Done() on a node that was never started would panic on a nil context.
	n.mu.RLock()
	ctx := n.ctx
	n.mu.RUnlock()

	if ctx == nil {
		return fmt.Errorf("node %s is not running", n.ID)
	}

	select {
	case n.inputChan <- msg:
		return nil
	case <-ctx.Done():
		return fmt.Errorf("node %s is stopped", n.ID)
	default:
		return fmt.Errorf("node %s input buffer is full", n.ID)
	}
}

// Connect connects this node's default output (port 0) to another node's input.
func (n *Node) Connect(targetNode *Node) {
	n.ConnectPort(targetNode, 0)
}

// ConnectPort connects one of this node's output ports to another node's input.
// Multi-output nodes such as switch and if use the port to route selectively;
// see Message.Ports.
func (n *Node) ConnectPort(targetNode *Node, port int) {
	n.mu.Lock()
	n.outputChans = append(n.outputChans, outputLink{ch: targetNode.inputChan, port: port})
	n.Outputs = append(n.Outputs, targetNode.ID)
	n.mu.Unlock()

	// The target's fields belong to the target's mutex, not ours. Taken as a
	// separate critical section so the two locks are never held at once.
	targetNode.mu.Lock()
	targetNode.Inputs = append(targetNode.Inputs, n.ID)
	targetNode.mu.Unlock()
}

// process handles incoming messages
func (n *Node) process() {
	for {
		select {
		case <-n.ctx.Done():
			return
		case msg := <-n.inputChan:
			n.handleMessage(msg)
		}
	}
}

// SetExecutionCallback sets a callback that fires after each execution
func (n *Node) SetExecutionCallback(cb ExecutionCallback) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.onExecution = cb
}

// execute runs the node's executor, converting a panic into an ordinary node
// error. Each node processes messages on its own goroutine, so an unrecovered
// panic in one executor would terminate the whole EdgeFlow process rather than
// just failing that node.
func (n *Node) execute(msg Message) (result Message, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("node %s (%s) panicked: %v", n.Name, n.Type, r)
			result = Message{}
		}
	}()
	return n.executor.Execute(n.ctx, msg)
}

// handleMessage processes a single message
func (n *Node) handleMessage(msg Message) {
	startTime := time.Now()

	// Executors index and assign into the payload directly, so a nil map would
	// panic on the first write ("assignment to entry in nil map"). A message can
	// legitimately arrive with no payload -- a Send from the API, or an upstream
	// node returning a bare Message -- so normalise it once here rather than
	// requiring every executor to guard.
	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}

	// Snapshot the input before executing: executors mutate the payload map in
	// place and return it as their output, so without a copy the reported Input
	// and Output are the same object and the debug view shows the post-execution
	// state for both. Only taken when something is listening, to avoid the copy
	// on every message when no debugger is attached.
	n.mu.RLock()
	inputCb := n.onExecution
	n.mu.RUnlock()
	var inputSnapshot map[string]interface{}
	if inputCb != nil {
		inputSnapshot = ClonePayload(msg.Payload)
	}

	// Execute node logic
	result, err := n.execute(msg)

	elapsed := time.Since(startTime).Milliseconds()

	if err != nil {
		n.mu.Lock()
		n.Status = NodeStatusError
		cb := n.onExecution
		n.mu.Unlock()

		// Emit execution event
		if cb != nil {
			cb(ExecutionEvent{
				NodeID:        n.ID,
				NodeName:      n.Name,
				NodeType:      n.Type,
				Input:         inputSnapshot,
				Output:        nil,
				Status:        "error",
				Error:         err.Error(),
				ExecutionTime: elapsed,
				Timestamp:     time.Now().UnixMilli(),
			})
		}

		// Send error message to outputs
		errorMsg := Message{
			Type:  MessageTypeError,
			Error: err,
			Payload: map[string]interface{}{
				"node_id": n.ID,
				"error":   err.Error(),
			},
		}
		n.sendToOutputs(errorMsg)
		return
	}

	// Emit execution event
	n.mu.RLock()
	cb := n.onExecution
	n.mu.RUnlock()
	if cb != nil {
		cb(ExecutionEvent{
			NodeID:        n.ID,
			NodeName:      n.Name,
			NodeType:      n.Type,
			Input:         inputSnapshot,
			Output:        ClonePayload(result.Payload),
			Status:        "success",
			ExecutionTime: elapsed,
			Timestamp:     time.Now().UnixMilli(),
		})
	}

	// Send result to connected nodes
	n.sendToOutputs(result)
}

// ClonePayload returns a deep copy of a message payload.
//
// Every recipient of a message must own its payload outright: nodes mutate the
// payload map in place, and each node runs in its own goroutine, so handing the
// same map to two nodes is an unsynchronized concurrent map write that crashes
// the process ("fatal error: concurrent map writes").
//
// Maps and slices are copied recursively. Scalars are treated as immutable and
// shared as-is, which preserves their concrete Go types (a JSON round-trip would
// silently rewrite every int as a float64).
func ClonePayload(v map[string]interface{}) map[string]interface{} {
	if v == nil {
		return nil
	}
	out := make(map[string]interface{}, len(v))
	for k, val := range v {
		out[k] = cloneValue(val)
	}
	return out
}

func cloneValue(v interface{}) interface{} {
	switch t := v.(type) {
	case map[string]interface{}:
		return ClonePayload(t)
	case []interface{}:
		s := make([]interface{}, len(t))
		for i, e := range t {
			s[i] = cloneValue(e)
		}
		return s
	case []byte:
		b := make([]byte, len(t))
		copy(b, t)
		return b
	default:
		return v
	}
}

// sendToOutputs broadcasts a message to all connected output nodes.
// Each recipient receives its own deep copy of the payload so that sibling
// branches cannot corrupt one another's data or race on the same map.
func (n *Node) sendToOutputs(msg Message) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, link := range n.outputChans {
		if !msg.routesTo(link.port) {
			continue
		}
		out := msg
		out.Payload = ClonePayload(msg.Payload)
		// The routing decision belongs to the sending node; it must not leak
		// downstream and constrain the next node's own outputs.
		out.Ports = nil
		select {
		case link.ch <- out:
		case <-n.ctx.Done():
			return
		default:
			// Output buffer full, skip
		}
	}
}

// routesTo reports whether the message should be delivered on the given port.
// A nil Ports slice means "no routing decision was made", which broadcasts.
func (m Message) routesTo(port int) bool {
	if m.Ports == nil {
		return true
	}
	for _, p := range m.Ports {
		if p == port {
			return true
		}
	}
	return false
}

// GetStatus returns the current node status
func (n *Node) GetStatus() NodeStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Status
}

// UpdateConfig updates the node configuration
func (n *Node) UpdateConfig(config map[string]interface{}) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Config = config

	// Re-initialize if running
	if n.Status == NodeStatusRunning {
		return n.executor.Init(config)
	}

	return nil
}
