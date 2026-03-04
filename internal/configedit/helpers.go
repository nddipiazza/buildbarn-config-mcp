// Package configedit provides stateless jsonnet config mutation helpers.
// These are used by both the standalone MCP tools and the HTTP handlers.
package configedit

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/google/go-jsonnet"
)

// EvalJsonnet evaluates a jsonnet snippet to a Go value.
// The VM is hardened: no file imports allowed, reduced max stack.
func EvalJsonnet(source string) (interface{}, error) {
	vm := jsonnet.MakeVM()
	vm.Importer(&jsonnet.MemoryImporter{}) // disable file/stdlib imports
	vm.MaxStack = 100
	jsonStr, err := vm.EvaluateAnonymousSnippet("config.jsonnet", source)
	if err != nil {
		return nil, fmt.Errorf("jsonnet eval: %v", err)
	}
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("JSON unmarshal: %v", err)
	}
	return data, nil
}

// DataToJsonnet converts a Go value to pretty-printed JSON (valid jsonnet).
func DataToJsonnet(data interface{}) (string, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// ParsePath decodes a JSON-encoded path array like `["a","0","b"]`.
func ParsePath(raw string) ([]string, error) {
	if raw == "" {
		return nil, fmt.Errorf("path is required")
	}
	var parts []interface{}
	if err := json.Unmarshal([]byte(raw), &parts); err != nil {
		return nil, fmt.Errorf("path must be a JSON array: %v", err)
	}
	result := make([]string, len(parts))
	for i, p := range parts {
		switch v := p.(type) {
		case string:
			result[i] = v
		case float64:
			result[i] = strconv.Itoa(int(v))
		default:
			result[i] = fmt.Sprintf("%v", v)
		}
	}
	return result, nil
}

// GetAt traverses data along path and returns the node.
func GetAt(data interface{}, path []string) (interface{}, error) {
	cur := data
	for i, seg := range path {
		switch v := cur.(type) {
		case map[string]interface{}:
			val, ok := v[seg]
			if !ok {
				return nil, fmt.Errorf("key %q not found at path[%d]", seg, i)
			}
			cur = val
		case []interface{}:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil, fmt.Errorf("invalid array index %q at path[%d]", seg, i)
			}
			cur = v[idx]
		default:
			return nil, fmt.Errorf("cannot traverse into %T at path[%d]=%q", cur, i, seg)
		}
	}
	return cur, nil
}

// SetAt returns a new data tree with the value at path replaced by newVal.
// Intermediate nil values are auto-created as empty maps.
func SetAt(data interface{}, path []string, newVal interface{}) (interface{}, error) {
	if len(path) == 0 {
		return newVal, nil
	}
	seg := path[0]
	rest := path[1:]
	if data == nil {
		data = map[string]interface{}{}
	}
	switch v := data.(type) {
	case map[string]interface{}:
		clone := make(map[string]interface{}, len(v))
		for k, val := range v {
			clone[k] = val
		}
		child := clone[seg]
		updated, err := SetAt(child, rest, newVal)
		if err != nil {
			return nil, err
		}
		clone[seg] = updated
		return clone, nil
	case []interface{}:
		idx, err := strconv.Atoi(seg)
		if err != nil || idx < 0 || idx >= len(v) {
			return nil, fmt.Errorf("invalid array index %q", seg)
		}
		clone := make([]interface{}, len(v))
		copy(clone, v)
		updated, err := SetAt(clone[idx], rest, newVal)
		if err != nil {
			return nil, err
		}
		clone[idx] = updated
		return clone, nil
	default:
		return nil, fmt.Errorf("cannot traverse into %T at segment %q", data, seg)
	}
}

// DeleteAt removes the node at path.
func DeleteAt(data interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return nil, fmt.Errorf("cannot delete root")
	}
	if len(path) == 1 {
		seg := path[0]
		switch v := data.(type) {
		case map[string]interface{}:
			clone := make(map[string]interface{}, len(v))
			for k, val := range v {
				clone[k] = val
			}
			delete(clone, seg)
			return clone, nil
		case []interface{}:
			idx, err := strconv.Atoi(seg)
			if err != nil || idx < 0 || idx >= len(v) {
				return nil, fmt.Errorf("invalid array index %q", seg)
			}
			clone := make([]interface{}, 0, len(v)-1)
			clone = append(clone, v[:idx]...)
			clone = append(clone, v[idx+1:]...)
			return clone, nil
		default:
			return nil, fmt.Errorf("cannot delete from %T", data)
		}
	}
	seg := path[0]
	rest := path[1:]
	switch v := data.(type) {
	case map[string]interface{}:
		clone := make(map[string]interface{}, len(v))
		for k, val := range v {
			clone[k] = val
		}
		child, ok := clone[seg]
		if !ok {
			return nil, fmt.Errorf("key %q not found", seg)
		}
		updated, err := DeleteAt(child, rest)
		if err != nil {
			return nil, err
		}
		clone[seg] = updated
		return clone, nil
	case []interface{}:
		idx, err := strconv.Atoi(seg)
		if err != nil || idx < 0 || idx >= len(v) {
			return nil, fmt.Errorf("invalid array index %q", seg)
		}
		clone := make([]interface{}, len(v))
		copy(clone, v)
		updated, err := DeleteAt(clone[idx], rest)
		if err != nil {
			return nil, err
		}
		clone[idx] = updated
		return clone, nil
	default:
		return nil, fmt.Errorf("cannot traverse into %T at segment %q", data, seg)
	}
}
