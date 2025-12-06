package test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func ParseJSON(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	var data map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &data)
	if err != nil {
		t.Fatal("JSON parse error:", err)
	}
	return data
}
