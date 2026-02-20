package out

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

// --- PrintJSON ---

func TestPrintJSON_validStruct(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	var buf bytes.Buffer
	err := PrintJSON(&buf, payload{Name: "Zurich", Age: 42})
	if err != nil {
		t.Fatalf("PrintJSON() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, `"name"`) {
		t.Errorf("output missing key %q: %s", "name", out)
	}
	if !strings.Contains(out, `"Zurich"`) {
		t.Errorf("output missing value %q: %s", "Zurich", out)
	}
	if !strings.Contains(out, `"age"`) {
		t.Errorf("output missing key %q: %s", "age", out)
	}
}

func TestPrintJSON_slice(t *testing.T) {
	var buf bytes.Buffer
	err := PrintJSON(&buf, []string{"rain", "snow"})
	if err != nil {
		t.Fatalf("PrintJSON() error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "rain") || !strings.Contains(out, "snow") {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestPrintJSON_emptySlice(t *testing.T) {
	var buf bytes.Buffer
	err := PrintJSON(&buf, []string{})
	if err != nil {
		t.Fatalf("PrintJSON() error: %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "[]" {
		t.Errorf("PrintJSON(empty slice) = %q, want %q", got, "[]")
	}
}

func TestPrintJSON_isIndented(t *testing.T) {
	type kv struct {
		Key string `json:"key"`
	}
	var buf bytes.Buffer
	_ = PrintJSON(&buf, kv{Key: "v"})
	// Indented output contains newlines.
	if !strings.Contains(buf.String(), "\n") {
		t.Error("PrintJSON() output is not indented (no newlines)")
	}
}

// --- WriteError ---

func TestWriteError_plainText(t *testing.T) {
	var buf bytes.Buffer
	err := WriteError(&buf, false, errors.New("something went wrong"))
	if err != nil {
		t.Fatalf("WriteError() returned unexpected error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, "something went wrong") {
		t.Errorf("output %q does not contain error message", got)
	}
	if !strings.HasPrefix(got, "Error:") {
		t.Errorf("plain text output %q should start with 'Error:'", got)
	}
}

func TestWriteError_jsonMode(t *testing.T) {
	var buf bytes.Buffer
	err := WriteError(&buf, true, errors.New("api unavailable"))
	if err != nil {
		t.Fatalf("WriteError() returned unexpected error: %v", err)
	}
	got := buf.String()
	if !strings.Contains(got, `"error"`) {
		t.Errorf("JSON output %q missing \"error\" key", got)
	}
	if !strings.Contains(got, "api unavailable") {
		t.Errorf("JSON output %q missing error message", got)
	}
}
