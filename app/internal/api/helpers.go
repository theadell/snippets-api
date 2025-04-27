package api

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type envelope map[string]any

const maxBodySize = 20 * 1_024 * 1_024 // 20MB

func readJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(dst)
	if err != nil {
		var unmarshalTypeError *json.UnmarshalTypeError
		switch {

		case errors.As(err, &unmarshalTypeError):
			return errors.New("incorrect JSON type for field " + unmarshalTypeError.Field)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("unexpected end of JSON input")

		case errors.Is(err, io.EOF):
			return errors.New("request body cannot be empty")

		default:
			return errors.New("invalid JSON provided")
		}
	}

	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func ok(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}

// StringValue returns the value of a string pointer or a default if nil
func stringValue(s *string, defaultVal string) string {
	if s != nil {
		return *s
	}
	return defaultVal
}

func toNullString(s *string) sql.NullString {
	if s != nil && *s != "" {
		return sql.NullString{String: *s, Valid: true}
	}
	return sql.NullString{Valid: false}
}

func writeError(w http.ResponseWriter, r *http.Request, status int, errorType string, message string) {
	response := Error{
		Status:  status,
		Error:   errorType,
		Message: message,
		Path:    &r.URL.Path,
		Timestamp: func() *time.Time {
			v := time.Now().UTC()
			return &v
		}(),
	}

	err := writeJSON(w, status, response)
	if err != nil {
		slog.Error("failed to write error response", "error", err)
	}
}

func notFoundError(w http.ResponseWriter, r *http.Request, message string) {
	writeError(w, r, http.StatusNotFound, "Not Found", message)
}

func badRequestError(w http.ResponseWriter, r *http.Request, message string) {
	writeError(w, r, http.StatusBadRequest, "Bad Request", message)
}

func unauthorizedError(w http.ResponseWriter, r *http.Request, message string) {
	writeError(w, r, http.StatusUnauthorized, "Unauthorized", message)
}

func forbiddenError(w http.ResponseWriter, r *http.Request, message string) {
	writeError(w, r, http.StatusForbidden, "Forbidden", message)
}

func internalServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal server error", "error", err, "path", r.URL.Path)
	writeError(w, r, http.StatusInternalServerError, "Internal Server Error", "An unexpected error occurred")
}

var defaultExpiresInDuration = 365 * 24 * time.Hour

// parseExpiresIn returns a sql.NullTime set to the current time plus the given duration string.
// If v is nil, it uses the default duration of 365 days from now.
// Returns an error if the string can't be parsed as a valid duration (e.g. "2h", "15m").
func parseExpiresIn(v *string) (sql.NullTime, error) {
	if v == nil {
		return sql.NullTime{Time: time.Now().Add(defaultExpiresInDuration), Valid: true}, nil
	}

	durString := *v
	dur, err := time.ParseDuration(durString)
	if err != nil {
		return sql.NullTime{}, fmt.Errorf("duration must be a valid time.Duration string: %v", err)
	}
	return sql.NullTime{Time: time.Now().UTC().Add(dur), Valid: true}, nil
}

func generateEditToken() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		panic("crypto/rand is unavailable: " + err.Error())
	}
	return hex.EncodeToString(bytes)
}
