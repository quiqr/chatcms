package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decodeError(t *testing.T, body []byte) Error {
	t.Helper()

	var envelope struct {
		Error Error `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatalf("response body %q is not the error envelope: %v", body, err)
	}
	return envelope.Error
}

func TestWriteErrorShape(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteError(recorder, http.StatusTeapot, "brewing", "This pot makes no coffee.")

	if recorder.Code != http.StatusTeapot {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusTeapot)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want JSON", got)
	}

	apiErr := decodeError(t, recorder.Body.Bytes())
	if apiErr.Code != "brewing" {
		t.Errorf("code = %q, want brewing", apiErr.Code)
	}
	if apiErr.Message != "This pot makes no coffee." {
		t.Errorf("message = %q, want the message passed in", apiErr.Message)
	}
}

func TestWriteJSONWithoutBody(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteJSON(recorder, http.StatusNoContent, nil)

	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if recorder.Body.Len() != 0 {
		t.Errorf("body = %q, want empty", recorder.Body.String())
	}
}
