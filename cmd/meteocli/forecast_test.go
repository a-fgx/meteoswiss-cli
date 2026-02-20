package main

import "testing"

func TestTruncate_shortString(t *testing.T) {
	got := truncate("Sunny", 10)
	if got != "Sunny" {
		t.Errorf("truncate(%q, 10) = %q, want %q", "Sunny", got, "Sunny")
	}
}

func TestTruncate_exactLength(t *testing.T) {
	s := "1234567890"
	got := truncate(s, 10)
	if got != s {
		t.Errorf("truncate(%q, 10) = %q, want %q", s, got, s)
	}
}

func TestTruncate_overLength(t *testing.T) {
	got := truncate("Heavy thunderstorm expected tonight", 10)
	runes := []rune(got)
	if len(runes) != 10 {
		t.Errorf("truncate result has %d runes, want 10: %q", len(runes), got)
	}
	// Must end with the ellipsis character.
	if runes[len(runes)-1] != '…' {
		t.Errorf("truncated string %q should end with '…'", got)
	}
}

func TestTruncate_unicode(t *testing.T) {
	// Emoji are multi-byte but single rune; truncation should count runes.
	s := "☀️ Sunny day ahead in Zurich"
	got := truncate(s, 8)
	runes := []rune(got)
	if len(runes) != 8 {
		t.Errorf("truncate unicode result has %d runes, want 8: %q", len(runes), got)
	}
}

func TestTruncate_empty(t *testing.T) {
	got := truncate("", 10)
	if got != "" {
		t.Errorf("truncate(%q, 10) = %q, want %q", "", got, "")
	}
}
