package utils

import (
	"encoding/json"
	"log"
)

// ToJSONString converts any interface to a JSON string representation.
// It handles potential errors by logging them and returning an empty string or error string,
// ensuring the application flow isn't disrupted by logging failures.
func ToJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("Failed to marshal object to JSON string: %v", err)
		return ""
	}
	return string(b)
}
