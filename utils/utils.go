package utils

import (
	"bytes"
	"encoding/gob"
)

// GetBytes ...
// Get byte array from any interface
func GetBytes(key interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(key)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
