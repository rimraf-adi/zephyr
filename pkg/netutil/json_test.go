package netutil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecodeJSONResponse_Success(t *testing.T) {
	obj := map[string]string{"foo": "bar"}
	data, _ := json.Marshal(obj)
	req := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()
	rw.Write(data)
	resp := rw.Result()
	resp.StatusCode = http.StatusOK
	var out map[string]string
	resp.Body = ioutil.NopCloser(bytes.NewReader(data))
	if err := DecodeJSONResponse(resp, &out); err != nil {
		t.Fatalf("DecodeJSONResponse failed: %v", err)
	}
	if out["foo"] != "bar" {
		t.Errorf("Expected foo=bar, got %+v", out)
	}
}

func TestDecodeJSONResponse_HTTPError(t *testing.T) {
	resp := httptest.NewRecorder().Result()
	resp.StatusCode = http.StatusBadRequest
	resp.Body = ioutil.NopCloser(bytes.NewReader([]byte("bad request")))
	var out map[string]string
	if err := DecodeJSONResponse(resp, &out); err == nil {
		t.Error("Expected error for HTTP error, got nil")
	}
}

func TestPrettyPrintJSON(t *testing.T) {
	obj := map[string]string{"foo": "bar"}
	str, err := PrettyPrintJSON(obj)
	if err != nil || str == "" {
		t.Errorf("PrettyPrintJSON failed: %v, %s", err, str)
	}
}

func TestValidateJSON(t *testing.T) {
	good := []byte(`{"foo": "bar"}`)
	bad := []byte(`{"foo": bar}`)
	if err := ValidateJSON(good); err != nil {
		t.Errorf("ValidateJSON failed for good JSON: %v", err)
	}
	if err := ValidateJSON(bad); err == nil {
		t.Error("Expected error for bad JSON, got nil")
	}
}

func TestJSONMapMethods(t *testing.T) {
	m := JSONMap{"foo": "bar", "num": 42, "bool": true, "arr": []interface{}{1, 2}, "map": map[string]interface{}{"x": 1}}
	if v, _ := m.GetString("foo"); v != "bar" {
		t.Error("GetString failed")
	}
	if _, err := m.GetString("missing"); err == nil {
		t.Error("Expected error for missing key")
	}
	if n, _ := m.GetInt("num"); n != 42 {
		t.Error("GetInt failed")
	}
	if b, _ := m.GetBool("bool"); !b {
		t.Error("GetBool failed")
	}
	if arr, _ := m.GetArray("arr"); len(arr) != 2 {
		t.Error("GetArray failed")
	}
	if mp, _ := m.GetMap("map"); mp["x"] != 1.0 {
		t.Error("GetMap failed")
	}
}

func TestJSONArrayMethods(t *testing.T) {
	a := JSONArray{"foo", 42, map[string]interface{}{"x": 1}}
	if v, _ := a.GetString(0); v != "foo" {
		t.Error("GetString failed")
	}
	if n, _ := a.GetInt(1); n != 42 {
		t.Error("GetInt failed")
	}
	if mp, _ := a.GetMap(2); mp["x"] != 1.0 {
		t.Error("GetMap failed")
	}
} 