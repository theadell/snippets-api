package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_StringValue(t *testing.T) {
	def := "default"
	a := "hi"
	tests := []struct {
		in   *string
		want string
	}{
		{&a, "hi"},
		{nil, def},
	}
	for _, tt := range tests {
		got := stringValue(tt.in, def)
		if got != tt.want {
			t.Errorf("stringValue(%v, %q) = %q; want %q", tt.in, def, got, tt.want)
		}
	}
}

func Test_readJSON(t *testing.T) {
	type Data struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	tests := []struct {
		name     string
		body     string
		want     Data
		wantErr  bool
		errorMsg string
	}{
		{
			name: "valid json",
			body: `{"name":"Alice","age":30}`,
			want: Data{Name: "Alice", Age: 30},
		},
		{
			name:     "empty body",
			body:     ``,
			wantErr:  true,
			errorMsg: "request body cannot be empty",
		},
		{
			name:     "invalid json",
			body:     `{name:}`,
			wantErr:  true,
			errorMsg: "invalid JSON provided",
		},
		{
			name:     "wrong type",
			body:     `{"name":"Bob","age":"not-an-int"}`,
			wantErr:  true,
			errorMsg: "incorrect JSON type",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			var got Data
			err := readJSON(rr, req, &got)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				if tt.errorMsg != "" && (err == nil || !errors.Is(err, io.EOF)) && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("got error %q, want error containing %q", err, tt.errorMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if got != tt.want {
					t.Errorf("got %+v, want %+v", got, tt.want)
				}
			}
		})
	}
}

func contains(target, el string) bool {
	return bytes.Contains([]byte(target), []byte(el))
}

func Test_writeJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	data := map[string]any{"hello": "world"}
	err := writeJSON(rr, http.StatusAccepted, data)
	if err != nil {
		t.Fatalf("writeJSON error: %v", err)
	}
	resp := rr.Result()
	if resp.StatusCode != http.StatusAccepted {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusAccepted)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}
	var decoded map[string]any
	json.NewDecoder(resp.Body).Decode(&decoded)
	if decoded["hello"] != "world" {
		t.Errorf("got body: %v", decoded)
	}
}

func Test_writeError(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/testpath", nil)
	writeError(rr, req, http.StatusBadRequest, "Bad Request", "This is a test error")
	resp := rr.Result()
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	if result["error"] != "Bad Request" {
		t.Errorf("error = %v, want %v", result["error"], "Bad Request")
	}
	if result["message"] != "This is a test error" {
		t.Errorf("message = %v, want %v", result["message"], "This is a test error")
	}
}

func Test_parseExpiresIn(t *testing.T) {
	str := "2h"
	nt, err := parseExpiresIn(&str)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !nt.Valid {
		t.Errorf("expected valid")
	}
	invalid := "notaduration"
	_, err = parseExpiresIn(&invalid)
	if err == nil {
		t.Errorf("expected error for invalid duration")
	}
}
