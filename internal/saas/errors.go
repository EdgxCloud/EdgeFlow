package saas

import "fmt"

// Common errors
type SaaSError struct {
	Code    string
	Message string
	Err     error
}

func (e *SaaSError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *SaaSError) Unwrap() error {
	return e.Err
}

// Error constructors
func ErrInvalidConfig(msg string) error {
	return &SaaSError{Code: "INVALID_CONFIG", Message: msg}
}

func ErrNotProvisioned(msg string) error {
	return &SaaSError{Code: "NOT_PROVISIONED", Message: msg}
}

func ErrConnectionFailed(err error) error {
	return &SaaSError{Code: "CONNECTION_FAILED", Message: "failed to connect to SaaS", Err: err}
}

func ErrAuthenticationFailed(msg string) error {
	return &SaaSError{Code: "AUTH_FAILED", Message: msg}
}

func ErrProvisioningFailed(err error) error {
	return &SaaSError{Code: "PROVISIONING_FAILED", Message: "device provisioning failed", Err: err}
}

func ErrCommandFailed(msg string, err error) error {
	return &SaaSError{Code: "COMMAND_FAILED", Message: msg, Err: err}
}

func ErrCommandTimeout(cmdID string) error {
	return &SaaSError{Code: "COMMAND_TIMEOUT", Message: fmt.Sprintf("command %s timed out", cmdID)}
}
