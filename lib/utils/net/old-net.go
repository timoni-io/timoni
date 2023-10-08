package net

import (
	"bytes"
	"io"
	"net/http"
)

// ReadBodyFromRequest ...
func ReadBodyFromRequest(r *http.Request) string {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return string(body)
}

// ReadBodyFromResponse ...
func ReadBodyFromResponse(r *http.Response) string {
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return string(body)
}
