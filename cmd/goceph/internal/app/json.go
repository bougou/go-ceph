package app

import (
	"encoding/json"
	"fmt"
	"io"
)

// WriteJSON encodes v as JSON to w.
func WriteJSON(w io.Writer, v any, pretty bool) error {
	enc := json.NewEncoder(w)
	if pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(v); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	return nil
}
