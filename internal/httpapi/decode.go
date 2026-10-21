package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

const MaxBodyBytes = 1 << 20

var (
	ErrMalformedBody = errors.New("request body is not valid JSON")
	ErrUnknownField  = errors.New("request body has an unknown field")
	ErrBodyTooLarge  = errors.New("request body is too large")
)

func Decode(w http.ResponseWriter, r *http.Request, target any) error {
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(target); err != nil {
		return classify(err)
	}
	if decoder.More() {
		return ErrMalformedBody
	}
	return nil
}

func classify(err error) error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return ErrBodyTooLarge
	}
	if strings.Contains(err.Error(), "unknown field") {
		return ErrUnknownField
	}
	return ErrMalformedBody
}

func WriteDecodeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrBodyTooLarge):
		WriteError(w, http.StatusRequestEntityTooLarge, CodePayloadTooLarge,
			"The request body is too large.")
	case errors.Is(err, ErrUnknownField):
		WriteError(w, http.StatusBadRequest, CodeBadRequest,
			"The request body has a field this endpoint does not accept.")
	default:
		WriteError(w, http.StatusBadRequest, CodeBadRequest,
			"The request body is not valid JSON.")
	}
}
