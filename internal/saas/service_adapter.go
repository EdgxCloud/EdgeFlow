package saas

// ServiceAdapter wraps api.Service to implement FlowService interface
// This avoids changing the api.Service signature to return 'any'
type ServiceAdapter struct {
	service interface{} // Will hold *api.Service
}

// NewServiceAdapter creates an adapter
func NewServiceAdapter(service interface{}) FlowService {
	return &ServiceAdapter{service: service}
}

// GetFlow implements FlowService.GetFlow
func (a *ServiceAdapter) GetFlow(id string) (any, error) {
	// Use reflection or type assertion to call the underlying method
	type getter interface {
		GetFlow(id string) (any, error)
	}
	if g, ok := a.service.(getter); ok {
		return g.GetFlow(id)
	}

	// Fallback: try different signature
	type getterEngine interface {
		GetFlow(id string) (interface{}, error)
	}
	if g, ok := a.service.(getterEngine); ok {
		return g.GetFlow(id)
	}

	return nil, ErrInvalidConfig("service does not implement GetFlow")
}

// StartFlow implements FlowService.StartFlow
func (a *ServiceAdapter) StartFlow(id string) error {
	type starter interface {
		StartFlow(id string) error
	}
	if s, ok := a.service.(starter); ok {
		return s.StartFlow(id)
	}
	return ErrInvalidConfig("service does not implement StartFlow")
}

// StopFlow implements FlowService.StopFlow
func (a *ServiceAdapter) StopFlow(id string) error {
	type stopper interface {
		StopFlow(id string) error
	}
	if s, ok := a.service.(stopper); ok {
		return s.StopFlow(id)
	}
	return ErrInvalidConfig("service does not implement StopFlow")
}

// GetFlows implements FlowService.GetFlows
func (a *ServiceAdapter) GetFlows() ([]any, error) {
	type getter interface {
		GetFlows() ([]any, error)
	}
	if g, ok := a.service.(getter); ok {
		return g.GetFlows()
	}

	// Fallback: try []interface{}
	type getterInterface interface {
		GetFlows() ([]interface{}, error)
	}
	if g, ok := a.service.(getterInterface); ok {
		return g.GetFlows()
	}

	return nil, ErrInvalidConfig("service does not implement GetFlows")
}
