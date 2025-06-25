package netutil

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// DecodeJSONResponse decodes a JSON response from an HTTP response
func DecodeJSONResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s - %s", resp.StatusCode, resp.Status, string(body))
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}
	
	if err := json.Unmarshal(body, v); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	return nil
}

// FetchAndDecodeJSON fetches a URL and decodes the JSON response
func FetchAndDecodeJSON(client *http.Client, url string, v interface{}) error {
	req, err := CreatePyPIRequest("GET", url)
	if err != nil {
		return err
	}
	
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %w", err)
	}
	
	return DecodeJSONResponse(resp, v)
}

// PrettyPrintJSON pretty prints a JSON object
func PrettyPrintJSON(v interface{}) (string, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}
	
	return string(data), nil
}

// ValidateJSON validates that a byte slice contains valid JSON
func ValidateJSON(data []byte) error {
	var v interface{}
	return json.Unmarshal(data, &v)
}

// JSONMap represents a generic JSON object
type JSONMap map[string]interface{}

// GetString gets a string value from a JSON map
func (m JSONMap) GetString(key string) (string, error) {
	if val, exists := m[key]; exists {
		if str, ok := val.(string); ok {
			return str, nil
		}
		return "", fmt.Errorf("key %s is not a string", key)
	}
	return "", fmt.Errorf("key %s not found", key)
}

// GetInt gets an integer value from a JSON map
func (m JSONMap) GetInt(key string) (int, error) {
	if val, exists := m[key]; exists {
		// Handle both float64 (from JSON) and int
		switch v := val.(type) {
		case float64:
			return int(v), nil
		case int:
			return v, nil
		default:
			return 0, fmt.Errorf("key %s is not a number", key)
		}
	}
	return 0, fmt.Errorf("key %s not found", key)
}

// GetBool gets a boolean value from a JSON map
func (m JSONMap) GetBool(key string) (bool, error) {
	if val, exists := m[key]; exists {
		if b, ok := val.(bool); ok {
			return b, nil
		}
		return false, fmt.Errorf("key %s is not a boolean", key)
	}
	return false, fmt.Errorf("key %s not found", key)
}

// GetMap gets a nested map from a JSON map
func (m JSONMap) GetMap(key string) (JSONMap, error) {
	if val, exists := m[key]; exists {
		if mapVal, ok := val.(map[string]interface{}); ok {
			return JSONMap(mapVal), nil
		}
		return nil, fmt.Errorf("key %s is not a map", key)
	}
	return nil, fmt.Errorf("key %s not found", key)
}

// GetArray gets an array from a JSON map
func (m JSONMap) GetArray(key string) ([]interface{}, error) {
	if val, exists := m[key]; exists {
		if arr, ok := val.([]interface{}); ok {
			return arr, nil
		}
		return nil, fmt.Errorf("key %s is not an array", key)
	}
	return nil, fmt.Errorf("key %s not found", key)
}

// ParseJSONMap parses a JSON byte slice into a JSONMap
func ParseJSONMap(data []byte) (JSONMap, error) {
	var m JSONMap
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return m, nil
}

// JSONArray represents a generic JSON array
type JSONArray []interface{}

// GetString gets a string value from a JSON array
func (a JSONArray) GetString(index int) (string, error) {
	if index < 0 || index >= len(a) {
		return "", fmt.Errorf("index %d out of bounds", index)
	}
	
	if str, ok := a[index].(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("index %d is not a string", index)
}

// GetInt gets an integer value from a JSON array
func (a JSONArray) GetInt(index int) (int, error) {
	if index < 0 || index >= len(a) {
		return 0, fmt.Errorf("index %d out of bounds", index)
	}
	
	switch v := a[index].(type) {
	case float64:
		return int(v), nil
	case int:
		return v, nil
	default:
		return 0, fmt.Errorf("index %d is not a number", index)
	}
}

// GetMap gets a map from a JSON array
func (a JSONArray) GetMap(index int) (JSONMap, error) {
	if index < 0 || index >= len(a) {
		return nil, fmt.Errorf("index %d out of bounds", index)
	}
	
	if mapVal, ok := a[index].(map[string]interface{}); ok {
		return JSONMap(mapVal), nil
	}
	return nil, fmt.Errorf("index %d is not a map", index)
}

// ParseJSONArray parses a JSON byte slice into a JSONArray
func ParseJSONArray(data []byte) (JSONArray, error) {
	var arr JSONArray
	if err := json.Unmarshal(data, &arr); err != nil {
		return nil, fmt.Errorf("failed to parse JSON array: %w", err)
	}
	return arr, nil
} 