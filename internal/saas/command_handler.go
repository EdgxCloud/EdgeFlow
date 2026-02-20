package saas

import (
	"fmt"

	"go.uber.org/zap"
)

// EdgeFlowCommandHandler handles commands from SaaS to control EdgeFlow
type EdgeFlowCommandHandler struct {
	logger       *zap.Logger
	flowService  FlowService
	shadowMgr    *ShadowManager
	service      SystemService // For metrics, executions, GPIO
}

// FlowService interface for flow operations (implemented by internal/api/service.go)
type FlowService interface {
	GetFlow(id string) (any, error)
	StartFlow(id string) error
	StopFlow(id string) error
	GetFlows() ([]any, error)
}

// SystemService interface for system operations
type SystemService interface {
	GetSystemInfo() (map[string]interface{}, error)
	GetExecutions() ([]interface{}, error)
	GetGPIOState() (map[string]interface{}, error)
}

// NewEdgeFlowCommandHandler creates a command handler
func NewEdgeFlowCommandHandler(logger *zap.Logger, flowService FlowService, shadowMgr *ShadowManager) *EdgeFlowCommandHandler {
	return &EdgeFlowCommandHandler{
		logger:      logger,
		flowService: flowService,
		shadowMgr:   shadowMgr,
	}
}

// SetSystemService sets the system service for additional data access
func (h *EdgeFlowCommandHandler) SetSystemService(service SystemService) {
	h.service = service
}

// HandleCommand processes a command from SaaS
func (h *EdgeFlowCommandHandler) HandleCommand(cmd *TunnelMessage) (*TunnelMessage, error) {
	h.logger.Info("Processing command",
		zap.String("action", cmd.Action),
		zap.String("id", cmd.ID))

	switch cmd.Action {
	case "health_check":
		return h.handleHealthCheck(cmd)

	case "list_flows":
		return h.handleListFlows(cmd)

	case "get_flow":
		return h.handleGetFlow(cmd)

	case "start_flow":
		return h.handleStartFlow(cmd)

	case "stop_flow":
		return h.handleStopFlow(cmd)

	case "create_flow":
		return h.handleCreateFlow(cmd)

	case "update_flow":
		return h.handleUpdateFlow(cmd)

	case "delete_flow":
		return h.handleDeleteFlow(cmd)

	case "get_shadow":
		return h.handleGetShadow(cmd)

	case "update_desired":
		return h.handleUpdateDesired(cmd)

	case "get_system_metrics":
		return h.handleGetSystemMetrics(cmd)

	case "get_executions":
		return h.handleGetExecutions(cmd)

	case "get_gpio_state":
		return h.handleGetGPIOState(cmd)

	default:
		return nil, fmt.Errorf("unknown command action: %s", cmd.Action)
	}
}

// handleHealthCheck returns device health status
func (h *EdgeFlowCommandHandler) handleHealthCheck(cmd *TunnelMessage) (*TunnelMessage, error) {
	data := map[string]interface{}{
		"status":  "healthy",
		"version": "1.0.0",
		"uptime":  "TODO", // Can add actual uptime tracking
	}

	return &TunnelMessage{
		Status: "success",
		Data:   data,
	}, nil
}

// handleListFlows returns all flows on this device
func (h *EdgeFlowCommandHandler) handleListFlows(cmd *TunnelMessage) (*TunnelMessage, error) {
	flows, err := h.flowService.GetFlows()
	if err != nil {
		return nil, err
	}

	return &TunnelMessage{
		Status: "success",
		Data:   map[string]interface{}{"flows": flows},
	}, nil
}

// handleGetFlow returns a specific flow
func (h *EdgeFlowCommandHandler) handleGetFlow(cmd *TunnelMessage) (*TunnelMessage, error) {
	flowID, ok := cmd.Payload["flow_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing flow_id")
	}

	flow, err := h.flowService.GetFlow(flowID)
	if err != nil {
		return nil, err
	}

	return &TunnelMessage{
		Status: "success",
		Data:   flow,
	}, nil
}

// handleStartFlow starts a flow execution
func (h *EdgeFlowCommandHandler) handleStartFlow(cmd *TunnelMessage) (*TunnelMessage, error) {
	flowID, ok := cmd.Payload["flow_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing flow_id")
	}

	if err := h.flowService.StartFlow(flowID); err != nil {
		return nil, err
	}

	return &TunnelMessage{
		Status: "success",
		Data:   map[string]interface{}{"flow_id": flowID, "status": "started"},
	}, nil
}

// handleStopFlow stops a flow execution
func (h *EdgeFlowCommandHandler) handleStopFlow(cmd *TunnelMessage) (*TunnelMessage, error) {
	flowID, ok := cmd.Payload["flow_id"].(string)
	if !ok {
		return nil, fmt.Errorf("missing flow_id")
	}

	if err := h.flowService.StopFlow(flowID); err != nil {
		return nil, err
	}

	return &TunnelMessage{
		Status: "success",
		Data:   map[string]interface{}{"flow_id": flowID, "status": "stopped"},
	}, nil
}

// handleCreateFlow creates a new flow (placeholder - would need CreateFlow method)
func (h *EdgeFlowCommandHandler) handleCreateFlow(cmd *TunnelMessage) (*TunnelMessage, error) {
	return nil, fmt.Errorf("create_flow not yet implemented")
}

// handleUpdateFlow updates a flow (placeholder - would need UpdateFlow method)
func (h *EdgeFlowCommandHandler) handleUpdateFlow(cmd *TunnelMessage) (*TunnelMessage, error) {
	return nil, fmt.Errorf("update_flow not yet implemented")
}

// handleDeleteFlow deletes a flow (placeholder - would need DeleteFlow method)
func (h *EdgeFlowCommandHandler) handleDeleteFlow(cmd *TunnelMessage) (*TunnelMessage, error) {
	return nil, fmt.Errorf("delete_flow not yet implemented")
}

// handleGetShadow returns current device shadow
func (h *EdgeFlowCommandHandler) handleGetShadow(cmd *TunnelMessage) (*TunnelMessage, error) {
	shadow, err := h.shadowMgr.GetShadow()
	if err != nil {
		return nil, err
	}

	return &TunnelMessage{
		Status: "success",
		Data:   shadow,
	}, nil
}

// handleUpdateDesired handles cloud-side updates to desired state
func (h *EdgeFlowCommandHandler) handleUpdateDesired(cmd *TunnelMessage) (*TunnelMessage, error) {
	desired, ok := cmd.Payload["desired"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid desired state payload")
	}

	h.logger.Info("Desired state updated by cloud", zap.Any("desired", desired))

	// Fetch latest shadow to get delta
	shadow, err := h.shadowMgr.GetShadow()
	if err != nil {
		return nil, err
	}

	return &TunnelMessage{
		Status: "success",
		Data: map[string]interface{}{
			"delta": shadow.Delta,
		},
	}, nil
}

// handleGetSystemMetrics returns system resource usage
func (h *EdgeFlowCommandHandler) handleGetSystemMetrics(cmd *TunnelMessage) (*TunnelMessage, error) {
	if h.service == nil {
		return nil, fmt.Errorf("system service not available")
	}

	metrics, err := h.service.GetSystemInfo()
	if err != nil {
		return nil, err
	}

	return &TunnelMessage{
		Status: "success",
		Data:   metrics,
	}, nil
}

// handleGetExecutions returns execution history
func (h *EdgeFlowCommandHandler) handleGetExecutions(cmd *TunnelMessage) (*TunnelMessage, error) {
	if h.service == nil {
		return nil, fmt.Errorf("system service not available")
	}

	executions, err := h.service.GetExecutions()
	if err != nil {
		return nil, err
	}

	return &TunnelMessage{
		Status: "success",
		Data:   map[string]interface{}{"executions": executions},
	}, nil
}

// handleGetGPIOState returns current GPIO pin states
func (h *EdgeFlowCommandHandler) handleGetGPIOState(cmd *TunnelMessage) (*TunnelMessage, error) {
	if h.service == nil {
		return nil, fmt.Errorf("system service not available")
	}

	gpioState, err := h.service.GetGPIOState()
	if err != nil {
		return nil, err
	}

	return &TunnelMessage{
		Status: "success",
		Data:   gpioState,
	}, nil
}
