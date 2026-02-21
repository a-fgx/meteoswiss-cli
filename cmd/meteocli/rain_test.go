package main

import (
	"testing"
	"time"

	"github.com/a-fgx/meteoswiss-cli/internal/api"
)

// anchor is a fixed reference time used across all table-driven tests so
// results are deterministic regardless of when the tests run.
var anchor = time.Date(2026, 2, 20, 12, 0, 0, 0, time.Local)

// makeGraph is a helper that builds a GraphData with 10-min high-res slots
// starting at anchor, optionally followed by hourly low-res slots.
//
//   - hiSlots: precipitation values for the 10-min phase
//   - loSlots: precipitation values for the 60-min phase (empty = no low-res)
func makeGraph(hiSlots, loSlots []float64) *api.GraphData {
	var loStart int64
	if len(loSlots) > 0 {
		loTime := anchor.Add(time.Duration(len(hiSlots)) * hiInterval)
		loStart = loTime.UnixMilli()
	}
	return &api.GraphData{
		Start:              anchor.UnixMilli(),
		StartLowResolution: loStart,
		Precipitation10m:   append([]float64{}, hiSlots...),
		Precipitation1h:    append([]float64{}, loSlots...),
	}
}

// --- graphRainInWindow ---

func TestGraphRainInWindow_noRain(t *testing.T) {
	// 12 dry 10-min slots covering [12:00, 13:50].
	g := makeGraph(make([]float64, 12), nil)
	// Query [12:00, 12:30] — should be dry.
	max, ok := graphRainInWindow(g, anchor, 30*time.Minute)
	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	if max != 0 {
		t.Errorf("maxMM = %.2f, want 0", max)
	}
}

func TestGraphRainInWindow_rainInWindow(t *testing.T) {
	// Slot at +20min (index 2) has 3.5 mm.
	slots := []float64{0, 0, 3.5, 0, 0, 0}
	g := makeGraph(slots, nil)
	max, ok := graphRainInWindow(g, anchor, 30*time.Minute)
	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	if max != 3.5 {
		t.Errorf("maxMM = %.2f, want 3.5", max)
	}
}

func TestGraphRainInWindow_rainOutsideWindow(t *testing.T) {
	// Rain is only at +60min (slot index 6), outside a 30-min window.
	slots := []float64{0, 0, 0, 0, 0, 0, 5.0}
	g := makeGraph(slots, nil)
	max, ok := graphRainInWindow(g, anchor, 30*time.Minute)
	if !ok {
		t.Fatal("expected ok=true, got false")
	}
	if max != 0 {
		t.Errorf("maxMM = %.2f, want 0 (rain is outside window)", max)
	}
}

func TestGraphRainInWindow_picksMaxNotFirst(t *testing.T) {
	// Multiple rainy slots; should return the maximum.
	slots := []float64{0, 1.0, 2.5, 0.5}
	g := makeGraph(slots, nil)
	max, ok := graphRainInWindow(g, anchor, 30*time.Minute)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if max != 2.5 {
		t.Errorf("maxMM = %.2f, want 2.5", max)
	}
}

func TestGraphRainInWindow_nowBeforeData(t *testing.T) {
	g := makeGraph([]float64{1.0, 2.0}, nil)
	// Query one hour before data starts — ok should be false.
	before := anchor.Add(-1 * time.Hour)
	_, ok := graphRainInWindow(g, before, 30*time.Minute)
	if ok {
		t.Error("expected ok=false when window is entirely before data, got true")
	}
}

func TestGraphRainInWindow_nowAfterData(t *testing.T) {
	// 3 high-res slots = 30 min of data ending at anchor+30min.
	g := makeGraph([]float64{0, 0, 0}, nil)
	after := anchor.Add(1 * time.Hour)
	_, ok := graphRainInWindow(g, after, 30*time.Minute)
	if ok {
		t.Error("expected ok=false when window is entirely after data, got true")
	}
}

func TestGraphRainInWindow_lowResSlot(t *testing.T) {
	// 3 high-res (10-min) dry slots, then 1 low-res (60-min) rainy slot.
	g := makeGraph([]float64{0, 0, 0}, []float64{4.0})
	// Query starting 30 min after anchor, within=60 min — hits the low-res slot.
	queryTime := anchor.Add(30 * time.Minute)
	max, ok := graphRainInWindow(g, queryTime, 60*time.Minute)
	if !ok {
		t.Fatal("expected ok=true")
	}
	if max != 4.0 {
		t.Errorf("maxMM = %.2f, want 4.0 (low-res rainy slot)", max)
	}
}

func TestGraphRainInWindow_emptyData(t *testing.T) {
	g := makeGraph([]float64{}, nil)
	_, ok := graphRainInWindow(g, anchor, 30*time.Minute)
	if ok {
		t.Error("expected ok=false for empty precipitation array")
	}
}

// --- checkRain ---

func TestCheckRain_graphDataNoRain(t *testing.T) {
	detail := &api.PLZDetail{
		Graph: makeGraph(make([]float64, 6), nil), // 6 dry 10-min slots
	}
	result := checkRain(8000, 30, detail, anchor)
	if result.RainExpected {
		t.Error("RainExpected = true, want false")
	}
	if result.MaxRainMM != 0 {
		t.Errorf("MaxRainMM = %.2f, want 0", result.MaxRainMM)
	}
}

func TestCheckRain_graphDataRainExpected(t *testing.T) {
	slots := []float64{0, 2.0, 0, 0}
	detail := &api.PLZDetail{
		Graph: makeGraph(slots, nil),
	}
	result := checkRain(8000, 30, detail, anchor)
	if !result.RainExpected {
		t.Error("RainExpected = false, want true")
	}
	if result.MaxRainMM != 2.0 {
		t.Errorf("MaxRainMM = %.2f, want 2.0", result.MaxRainMM)
	}
}

func TestCheckRain_fallbackToDailyRainy(t *testing.T) {
	// No graph data — falls back to Forecast.
	detail := &api.PLZDetail{
		Forecast: []api.DayForecast{
			{DayDate: "2026-02-20", Precipitation: 5.5},
		},
	}
	result := checkRain(8000, 30, detail, anchor)
	if !result.RainExpected {
		t.Error("RainExpected = false, want true")
	}
	if result.MaxRainMM != 5.5 {
		t.Errorf("MaxRainMM = %.2f, want 5.5", result.MaxRainMM)
	}
}

func TestCheckRain_fallbackToDailyDry(t *testing.T) {
	detail := &api.PLZDetail{
		Forecast: []api.DayForecast{
			{DayDate: "2026-02-20", Precipitation: 0},
		},
	}
	result := checkRain(8000, 30, detail, anchor)
	if result.RainExpected {
		t.Error("RainExpected = true, want false")
	}
}

func TestCheckRain_noDataAtAll(t *testing.T) {
	result := checkRain(8000, 30, &api.PLZDetail{}, anchor)
	if result.RainExpected {
		t.Error("RainExpected = true, want false for empty detail")
	}
	if result.Message == "" {
		t.Error("Message should not be empty")
	}
}

func TestCheckRain_metadataPassedThrough(t *testing.T) {
	detail := &api.PLZDetail{}
	result := checkRain(3000, 60, detail, anchor)
	if result.PLZ != 3000 {
		t.Errorf("PLZ = %d, want 3000", result.PLZ)
	}
	if result.WithinMinutes != 60 {
		t.Errorf("WithinMinutes = %d, want 60", result.WithinMinutes)
	}
}

// --- CLI validation ---

func TestExecute_rainMissingZip(t *testing.T) {
	err := execute([]string{"rain"})
	if err == nil {
		t.Error("expected error when --zip is missing, got nil")
	}
}

func TestExecute_rainInvalidPLZ(t *testing.T) {
	err := execute([]string{"rain", "--zip", "500"})
	if err == nil {
		t.Error("expected error for invalid PLZ 500, got nil")
	}
}

func TestExecute_rainWithinTooLow(t *testing.T) {
	err := execute([]string{"rain", "--zip", "8000", "--within", "0"})
	if err == nil {
		t.Error("expected error for --within 0, got nil")
	}
}

func TestExecute_rainWithinTooHigh(t *testing.T) {
	err := execute([]string{"rain", "--zip", "8000", "--within", "1441"})
	if err == nil {
		t.Error("expected error for --within 1441, got nil")
	}
}
