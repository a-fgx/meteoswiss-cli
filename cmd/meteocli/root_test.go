package main

import (
	"strings"
	"testing"
)

// --- requirePLZ ---

func TestRequirePLZ_valid(t *testing.T) {
	validCodes := []int{1000, 3000, 8000, 9000, 9999}
	for _, plz := range validCodes {
		if err := requirePLZ(plz); err != nil {
			t.Errorf("requirePLZ(%d) returned unexpected error: %v", plz, err)
		}
	}
}

func TestRequirePLZ_tooLow(t *testing.T) {
	invalidCodes := []int{0, 1, 500, 999}
	for _, plz := range invalidCodes {
		if err := requirePLZ(plz); err == nil {
			t.Errorf("requirePLZ(%d) expected error, got nil", plz)
		}
	}
}

func TestRequirePLZ_tooHigh(t *testing.T) {
	invalidCodes := []int{10000, 12345, 99999}
	for _, plz := range invalidCodes {
		if err := requirePLZ(plz); err == nil {
			t.Errorf("requirePLZ(%d) expected error, got nil", plz)
		}
	}
}

func TestRequirePLZ_errorMessage(t *testing.T) {
	err := requirePLZ(500)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error %q should mention the invalid code 500", err.Error())
	}
}

// --- execute: flag validation (no network calls) ---

func TestExecute_weatherMissingZip(t *testing.T) {
	// --zip is required; cobra returns an error before RunE.
	err := execute([]string{"weather"})
	if err == nil {
		t.Error("expected error when --zip is missing, got nil")
	}
}

func TestExecute_forecastInvalidDays(t *testing.T) {
	// --days 0 should fail validation inside RunE before any API call.
	// We need a valid PLZ so we get past requirePLZ, but days=0 should fail.
	// Because the API would be called we only check days > 10 / days < 1;
	// the test server is not running so the API call itself would error too —
	// but the days validation fires first.
	err := execute([]string{"forecast", "--zip", "8000", "--days", "0"})
	if err == nil {
		t.Error("expected error for --days 0, got nil")
	}
	if !strings.Contains(err.Error(), "--days") {
		t.Errorf("error %q should mention --days", err.Error())
	}
}

func TestExecute_forecastDaysTooHigh(t *testing.T) {
	err := execute([]string{"forecast", "--zip", "8000", "--days", "11"})
	if err == nil {
		t.Error("expected error for --days 11, got nil")
	}
}

func TestExecute_warningsInvalidMinLevel(t *testing.T) {
	for _, level := range []string{"0", "6"} {
		err := execute([]string{"warnings", "--min-level", level})
		if err == nil {
			t.Errorf("expected error for --min-level %s, got nil", level)
		}
		if !strings.Contains(err.Error(), "--min-level") {
			t.Errorf("error %q should mention --min-level", err.Error())
		}
	}
}

func TestExecute_weatherInvalidPLZ(t *testing.T) {
	for _, zip := range []string{"500", "10000"} {
		err := execute([]string{"weather", "--zip", zip})
		if err == nil {
			t.Errorf("expected error for --zip %s, got nil", zip)
		}
	}
}

func TestExecute_forecastInvalidPLZ(t *testing.T) {
	err := execute([]string{"forecast", "--zip", "999"})
	if err == nil {
		t.Error("expected error for invalid PLZ 999, got nil")
	}
}

func TestExecute_version(t *testing.T) {
	// --version should succeed with no error.
	err := execute([]string{"--version"})
	if err != nil {
		t.Errorf("execute(--version) unexpected error: %v", err)
	}
}

func TestExecute_help(t *testing.T) {
	err := execute([]string{"--help"})
	if err != nil {
		t.Errorf("execute(--help) unexpected error: %v", err)
	}
}
