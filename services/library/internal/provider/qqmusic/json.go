package qqmusic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
)

func decodeObject(raw []byte) (map[string]json.RawMessage, error) {
	if len(raw) == 0 || bytes.Equal(raw, []byte("null")) {
		return nil, fmt.Errorf("object is empty")
	}
	var object map[string]json.RawMessage
	if err := decodeJSON(raw, &object); err != nil || object == nil {
		return nil, fmt.Errorf("object is invalid")
	}
	return object, nil
}

func nestedObject(object map[string]json.RawMessage, keys ...string) (map[string]json.RawMessage, bool) {
	for _, key := range keys {
		if raw, ok := object[key]; ok {
			nested, err := decodeObject(raw)
			return nested, err == nil
		}
	}
	return nil, false
}

func stringValue(object map[string]json.RawMessage, keys ...string) string {
	for _, key := range keys {
		raw, ok := object[key]
		if !ok || bytes.Equal(raw, []byte("null")) {
			continue
		}
		var value string
		if json.Unmarshal(raw, &value) == nil {
			return value
		}
		var number json.Number
		if json.Unmarshal(raw, &number) == nil {
			return number.String()
		}
	}
	return ""
}

func int64Value(object map[string]json.RawMessage, keys ...string) int64 {
	for _, key := range keys {
		raw, ok := object[key]
		if !ok || bytes.Equal(raw, []byte("null")) {
			continue
		}
		var number json.Number
		if json.Unmarshal(raw, &number) == nil {
			value, err := number.Int64()
			if err == nil {
				return value
			}
		}
		var text string
		if json.Unmarshal(raw, &text) == nil {
			value, err := strconv.ParseInt(text, 10, 64)
			if err == nil {
				return value
			}
		}
	}
	return 0
}

func hasAnyField(object map[string]json.RawMessage, keys ...string) bool {
	for _, key := range keys {
		if _, ok := object[key]; ok {
			return true
		}
	}
	return false
}
