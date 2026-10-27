package observability

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestJSONLogger(t *testing.T) {
	var output bytes.Buffer

	logger := slog.New(slog.NewJSONHandler(
		&output,
		&slog.HandlerOptions{Level: slog.LevelInfo},
	))

	logger.Info(
		"request completed",
		"method", "POST",
		"path", "/graphql/portal",
		"status", 200,
	)

	var entry map[string]any

	if err := json.Unmarshal(output.Bytes(), &entry); err != nil {
		t.Fatalf("invalid JSON log: %v", err)
	}

	if entry["msg"] != "request completed" {
		t.Fatalf("unexpected message: %v", entry["msg"])
	}

	if entry["method"] != "POST" {
		t.Fatalf("unexpected method: %v", entry["method"])
	}

	if entry["status"] != float64(200) {
		t.Fatalf("unexpected status: %v", entry["status"])
	}
}
