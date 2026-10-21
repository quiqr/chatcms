package httpapi

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type payload struct {
	Title string `json:"title"`
}

func decodeRequest(t *testing.T, body string) error {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/api/thing", strings.NewReader(body))
	recorder := httptest.NewRecorder()

	var target payload
	return Decode(recorder, request, &target)
}

func TestDecodeAcceptsAValidBody(t *testing.T) {
	if err := decodeRequest(t, `{"title":"About us"}`); err != nil {
		t.Errorf("Decode() error = %v, want nil", err)
	}
}

func TestDecodeRejectsMalformedBody(t *testing.T) {
	err := decodeRequest(t, `{"title":`)
	if !errors.Is(err, ErrMalformedBody) {
		t.Errorf("Decode() error = %v, want ErrMalformedBody", err)
	}
}

func TestDecodeRejectsTrailingContent(t *testing.T) {
	err := decodeRequest(t, `{"title":"a"}{"title":"b"}`)
	if !errors.Is(err, ErrMalformedBody) {
		t.Errorf("Decode() error = %v, want ErrMalformedBody", err)
	}
}

func TestDecodeRejectsUnknownField(t *testing.T) {
	err := decodeRequest(t, `{"title":"About us","sneaky":true}`)
	if !errors.Is(err, ErrUnknownField) {
		t.Errorf("Decode() error = %v, want ErrUnknownField", err)
	}
}

func TestDecodeRejectsOversizedBody(t *testing.T) {
	huge := `{"title":"` + strings.Repeat("x", MaxBodyBytes+1) + `"}`

	err := decodeRequest(t, huge)
	if !errors.Is(err, ErrBodyTooLarge) {
		t.Errorf("Decode() error = %v, want ErrBodyTooLarge", err)
	}
}

func TestWriteDecodeErrorDoesNotEchoTheBody(t *testing.T) {
	cases := map[string]struct {
		err        error
		wantStatus int
	}{
		"malformed": {ErrMalformedBody, http.StatusBadRequest},
		"unknown":   {ErrUnknownField, http.StatusBadRequest},
		"too large": {ErrBodyTooLarge, http.StatusRequestEntityTooLarge},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			WriteDecodeError(recorder, tc.err)

			if recorder.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tc.wantStatus)
			}
			apiErr := decodeError(t, recorder.Body.Bytes())
			if strings.Contains(apiErr.Message, "{") {
				t.Errorf("message = %q, want no echo of the body", apiErr.Message)
			}
		})
	}
}
