package test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func ParseJSON(t *testing.T, rr *httptest.ResponseRecorder) map[string]interface{} {
	var data map[string]interface{}
	_ = json.Unmarshal(rr.Body.Bytes(), &data)
	return data
}

func ResetDB(t *testing.T) {
	_, err := TestDB.Exec(`
		TRUNCATE TABLE
		users,
		files,
		shared,
		download,
		usersLoginSession,
		jwt_blacklist
		CASCADE;
	`)
	if err != nil {
		t.Fatal(err)
	}
}
