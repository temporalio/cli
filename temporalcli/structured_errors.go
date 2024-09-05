package temporalcli

import "go.temporal.io/sdk/temporal"

type structuredError struct {
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Details any    `json:"details,omitempty"`
}

func fromApplicationError(err *temporal.ApplicationError) *structuredError {
	return &structuredError{
		Message: err.Error(),
		Type:    err.Type(),
		Details: err.Details(),
	}
}
