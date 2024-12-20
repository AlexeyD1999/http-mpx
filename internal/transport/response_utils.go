package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func Respond(w http.ResponseWriter, status int, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err = w.Write(data); err != nil {
		return fmt.Errorf("failed to write response: %w", err)
	}

	return nil
}
