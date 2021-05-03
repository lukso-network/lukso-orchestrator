package kv

import (
	"bytes"
	"encoding/json"
)

func encode(v interface{}) ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := json.NewEncoder(buffer).Encode(v); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func decode(data []byte, v interface{}) error {
	var buf bytes.Buffer
	if _, err := buf.Write(data); err != nil {
		return err
	}

	if err := json.NewDecoder(&buf).Decode(v); err != nil {
		return err
	}
	return nil
}
