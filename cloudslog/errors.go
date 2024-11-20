package cloudslog

import (
	"encoding/json"
	"log/slog"
)

// Errors creates a slog attribute for a list of errors.
// unlike slog.Any, this will render the error strings when using slog.JSONHandler.
func Errors(errors []error) slog.Attr {
	jsonErrors := make([]error, len(errors))
	for i, err := range errors {
		jsonErrors[i] = &jsonError{error: err}
	}
	return slog.Any("errors", jsonErrors)
}

type jsonError struct {
	error
}

func (j jsonError) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.error.Error())
}
