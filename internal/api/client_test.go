package api

import (
	"encoding/json"
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
			Time:        "2026-02-20T12:00:00",
			Icon:        1,
			Temperature: 5.5,
		},
		TenDaysForecast: []DayForecast{
			{
				DayDate:        "2026-02-20",
				IconDay:        2,
				TemperatureMax: 8.0,
				TemperatureMin: 2.0,
				Precipitation:  0.5,
				WindDirection:  270,
				WindSpeed:      15,
			},
		},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify the PLZ is converted to 6 digits.
		if got := r.URL.Query().Get("plz"); got != "800000" {
			t.Errorf("PLZ query param = %q, want %q", got, "800000")
		}
		// Verify User-Agent is set.
		if ua := r.Header.Get("User-Agent"); !strings.HasPrefix(ua, "meteocli/") {
			t.Errorf("User-Agent = %q, expected meteocli/ prefix", ua)
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
	if len(got.TenDaysForecast) != 1 {
		t.Fatalf("TenDaysForecast len = %d, want 1", len(got.TenDaysForecast))
	}
	if got.TenDaysForecast[0].DayDate != "2026-02-20" {
		t.Errorf("DayDate = %q, want %q", got.TenDaysForecast[0].DayDate, "2026-02-20")
	}
}

func TestPLZDetail_nonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.PLZDetail(8000)
	if err == nil {
		t.Fatal("expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error %q should mention 404", err.Error())
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

// --- Warnings ---

func TestWarnings_success(t *testing.T) {
	want := []Warning{
		{WarnType: 1, WarnLevel: 3, Headline: "Heavy thunderstorm expected", Regions: []string{"CH01", "CH02"}},
		{WarnType: 2, WarnLevel: 2, Headline: "Rain warning"},
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/warnings") {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(want)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	got, err := client.Warnings()
	if err != nil {
		t.Fatalf("Warnings() unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("Warnings() len = %d, want 2", len(got))
	}
	if got[0].Headline != want[0].Headline {
		t.Errorf("Headline = %q, want %q", got[0].Headline, want[0].Headline)
	}
}

func TestWarnings_emptyList(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	got, err := client.Warnings()
	if err != nil {
		t.Fatalf("Warnings() unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected empty slice, got %d items", len(got))
	}
}

func TestWarnings_nonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, err := client.Warnings()
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
}

// --- Accept header ---

func TestClient_acceptHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if accept := r.Header.Get("Accept"); accept != "application/json" {
			t.Errorf("Accept = %q, want %q", accept, "application/json")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()

	client := newTestClient(srv.URL)
	_, _ = client.Warnings()
}
