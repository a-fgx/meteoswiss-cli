package api

import "testing"

// --- IconDescription ---

func TestIconDescription_knownCodes(t *testing.T) {
	cases := []struct {
		code int
		want string
	}{
		{1, "Sunny"},
		{3, "Partly cloudy"},
		{10, "Thunderstorm"},
		{16, "Clear night"},
		{42, "Blowing snow"},
	}
	for _, tc := range cases {
		got := IconDescription(tc.code)
		if got != tc.want {
			t.Errorf("IconDescription(%d) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestIconDescription_unknownCode(t *testing.T) {
	for _, code := range []int{0, 43, -1, 999} {
		got := IconDescription(code)
		if got != "Unknown" {
			t.Errorf("IconDescription(%d) = %q, want %q", code, got, "Unknown")
		}
	}
}

// --- IconEmoji ---

func TestIconEmoji_knownCodes(t *testing.T) {
	cases := []struct {
		code int
		want string
	}{
		{1, "☀️"},
		{5, "☁️"},
		{10, "⛈️"},
		{12, "❄️"},
	}
	for _, tc := range cases {
		got := IconEmoji(tc.code)
		if got != tc.want {
			t.Errorf("IconEmoji(%d) = %q, want %q", tc.code, got, tc.want)
		}
	}
}

func TestIconEmoji_unknownCode(t *testing.T) {
	for _, code := range []int{0, 43, 100} {
		got := IconEmoji(code)
		if got != "?" {
			t.Errorf("IconEmoji(%d) = %q, want \"?\"", code, got)
		}
	}
}

// --- WindDirectionLabel ---

func TestWindDirectionLabel(t *testing.T) {
	cases := []struct {
		deg  int
		want string
	}{
		{0, "N"},
		{90, "E"},
		{180, "S"},
		{270, "W"},
		{360, "N"},   // wraps back to North
		{45, "NE"},
		{135, "SE"},
		{225, "SW"},
		{315, "NW"},
		{11, "N"},    // just inside N bucket
		{12, "NNE"},  // just inside NNE bucket
	}
	for _, tc := range cases {
		got := WindDirectionLabel(tc.deg)
		if got != tc.want {
			t.Errorf("WindDirectionLabel(%d) = %q, want %q", tc.deg, got, tc.want)
		}
	}
}

func TestWindDirectionLabel_negative(t *testing.T) {
	got := WindDirectionLabel(-1)
	if got != "—" {
		t.Errorf("WindDirectionLabel(-1) = %q, want \"—\"", got)
	}
}

// --- plz6 ---

func TestPlz6(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{8000, 800000},
		{3000, 300000},
		{1200, 120000},
		{1000, 100000},
		{9999, 999900},
	}
	for _, tc := range cases {
		got := plz6(tc.in)
		if got != tc.want {
			t.Errorf("plz6(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

// --- WeatherIcon completeness ---

func TestWeatherIconCoversAll42Codes(t *testing.T) {
	for i := 1; i <= 42; i++ {
		if _, ok := WeatherIcon[i]; !ok {
			t.Errorf("WeatherIcon missing entry for icon code %d", i)
		}
	}
}
