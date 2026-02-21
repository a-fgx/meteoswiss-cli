package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient returns a Client pointed at the given test server URL.
func newTestClient(serverURL string) *Client {
	return &Client{
		http:    &http.Client{},
		baseURL: serverURL,
	}
}

// --- PLZDetail ---

func TestPLZDetail_success(t *testing.T) {
	want := PLZDetail{
		CurrentWeather: CurrentWeather{
			Time:        1740052800, // 2026-02-20 12:00:00 UTC
			Icon:        1,
			Temperature: 5.5,
		},
		Forecast: []DayForecast{
			{
				DayDate:        "2026-02-20",
				IconDay:        2,
				TemperatureMax: 8.0,
				TemperatureMin: 2.0,
				Precipitation:  0.5,
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the PLZ is converted to 6 digits.
		if got := r.URL.Query().Get("plz"); got != "800000" {
			t.Errorf("PLZ query param = %q, want %q", got, "800000")
		}
		// Verify User-Agent is set.
		if ua := r.Header.Get("User-Agent"); !strings.HasPrefix(ua, "meteoswiss-cli/") {
			t.Errorf("User-Agent = %q, expected meteoswiss-cli/ prefix", ua)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	got, err := client.PLZDetail(8000)
	if err != nil {
		t.Fatalf("PLZDetail() unexpected error: %v", err)
	}

	if got.CurrentWeather.Temperature != want.CurrentWeather.Temperature {
		t.Errorf("Temperature = %.1f, want %.1f", got.CurrentWeather.Temperature, want.CurrentWeather.Temperature)
	}
	if got.CurrentWeather.Icon != want.CurrentWeather.Icon {
		t.Errorf("Icon = %d, want %d", got.CurrentWeather.Icon, want.CurrentWeather.Icon)
	}
	if len(got.Forecast) != 1 {
		t.Fatalf("Forecast len = %d, want 1", len(got.Forecast))
	}
	if got.Forecast[0].DayDate != "2026-02-20" {
		t.Errorf("DayDate = %q, want %q", got.Forecast[0].DayDate, "2026-02-20")
	}
}

// TestPLZDetail_integerTimestamps verifies that the client can decode a
// realistic API response where time fields are Unix timestamps (integers),
// not strings. This catches type mismatches between struct definitions and
// the actual API wire format.
func TestPLZDetail_integerTimestamps(t *testing.T) {
	const body = `{
		"currentWeather": {"time": 1740052800000, "icon": 3, "temperature": 7.2},
		"forecast": [],
		"graph": {
			"start": 1740052800000,
			"startLowResolution": 1740060000000,
			"precipitation10m": [0.0, 0.2, 0.0],
			"precipitation1h": [0.5, 1.2]
		}
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	got, err := client.PLZDetail(8000)
	if err != nil {
		t.Fatalf("PLZDetail() unexpected error: %v", err)
	}
	if got.CurrentWeather.Time != 1740052800000 {
		t.Errorf("CurrentWeather.Time = %d, want 1740052800000", got.CurrentWeather.Time)
	}
	if got.Graph == nil {
		t.Fatal("Graph is nil, want non-nil")
	}
	if got.Graph.Start != 1740052800000 {
		t.Errorf("Graph.Start = %d, want 1740052800000", got.Graph.Start)
	}
	if got.Graph.StartLowResolution != 1740060000000 {
		t.Errorf("Graph.StartLowResolution = %d, want 1740060000000", got.Graph.StartLowResolution)
	}
	if len(got.Graph.Precipitation10m) != 3 {
		t.Errorf("Precipitation10m len = %d, want 3", len(got.Graph.Precipitation10m))
	}
	if len(got.Graph.Precipitation1h) != 2 {
		t.Errorf("Precipitation1h len = %d, want 2", len(got.Graph.Precipitation1h))
	}
}

func TestPLZDetail_nonOKStatus(t *testing.T) {
	cases := []struct {
		name   string
		status int
	}{
		{"not found", http.StatusNotFound},
		// 500 is what the MeteoSwiss API returns for unsupported PLZ codes
		// (e.g. 8000 → 800000, a generic Zürich meta-code with no weather station).
		{"server error", http.StatusInternalServerError},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, http.StatusText(tc.status), tc.status)
			}))
			defer srv.Close()

			client := newTestClient(srv.URL)
			_, err := client.PLZDetail(8000)
			if err == nil {
				t.Fatalf("expected error for %d, got nil", tc.status)
			}
			statusStr := fmt.Sprintf("%d", tc.status)
			if !strings.Contains(err.Error(), statusStr) {
				t.Errorf("error %q should mention %s", err.Error(), statusStr)
			}
		})
	}
}

func TestPLZDetail_badJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.PLZDetail(8000)
	if err == nil {
		t.Fatal("expected error for bad JSON, got nil")
	}
}

func TestPLZDetail_serverDown(t *testing.T) {
	// Point at a server that isn't listening.
	client := newTestClient("http://127.0.0.1:1")
	_, err := client.PLZDetail(8000)
	if err == nil {
		t.Fatal("expected error for unreachable server, got nil")
	}
}

// TestPLZDetail_warnings verifies that warnings embedded in the plzDetail
// response are decoded correctly. Warnings are PLZ-specific; there is no
// standalone /warnings endpoint.
func TestPLZDetail_warnings(t *testing.T) {
	const body = `{
		"currentWeather": {"time": 1740052800000, "icon": 1, "temperature": 5.0},
		"forecast": [],
		"warnings": [
			{"warnType": 2, "warnLevel": 3, "headline": "Heavy rain expected"},
			{"warnType": 0, "warnLevel": 2, "headline": "Strong winds"}
		]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	got, err := client.PLZDetail(2555)
	if err != nil {
		t.Fatalf("PLZDetail() unexpected error: %v", err)
	}
	if len(got.Warnings) != 2 {
		t.Fatalf("Warnings len = %d, want 2", len(got.Warnings))
	}
	if got.Warnings[0].WarnLevel != 3 {
		t.Errorf("WarnLevel = %d, want 3", got.Warnings[0].WarnLevel)
	}
	if got.Warnings[0].Headline != "Heavy rain expected" {
		t.Errorf("Headline = %q, want \"Heavy rain expected\"", got.Warnings[0].Headline)
	}
}

// --- Accept header ---

func TestClient_acceptHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if accept := r.Header.Get("Accept"); accept != "application/json" {
			t.Errorf("Accept = %q, want %q", accept, "application/json")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, _ = client.PLZDetail(8000)
}
