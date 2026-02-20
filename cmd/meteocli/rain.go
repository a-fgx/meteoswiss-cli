package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/meteocli/internal/api"
	"github.com/user/meteocli/internal/out"
)

// rainResult is the structured result for the rain command.
type rainResult struct {
	PLZ           int     `json:"plz"`
	WithinMinutes int     `json:"within_minutes"`
	RainExpected  bool    `json:"rain_expected"`
	MaxRainMM     float64 `json:"max_rain_mm"`
	Message       string  `json:"message"`
}

func newRainCmd(flags *rootFlags) *cobra.Command {
	var plz int
	var within int

	cmd := &cobra.Command{
		Use:   "rain",
		Short: "Check if rain is expected within a time window for a Swiss postal code",
		Example: `  # Rain check with the default 30-minute window for Zurich
  meteocli rain --zip 8000

  # Rain check for the next 60 minutes in Bern
  meteocli rain --zip 3000 --within 60

  # As JSON
  meteocli rain --zip 8000 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePLZ(plz); err != nil {
				return err
			}
			if within < 1 || within > 1440 {
				return fmt.Errorf("--within must be between 1 and 1440 minutes")
			}

			client := api.New()
			detail, err := client.PLZDetail(plz)
			if err != nil {
				return err
			}

			result := checkRain(plz, within, detail, time.Now())

			if flags.asJSON {
				return out.PrintJSON(os.Stdout, result)
			}
			printRainCheck(result)
			return nil
		},
	}

	cmd.Flags().IntVar(&plz, "zip", 0, "Swiss postal code (e.g. 8000 for Zurich)")
	cmd.Flags().IntVar(&within, "within", 30, "look-ahead window in minutes (1–1440)")
	_ = cmd.MarkFlagRequired("zip")
	return cmd
}

// checkRain determines whether rain is expected in the next `within` minutes.
// It uses the high-resolution graph data when available (10-minute intervals
// up to StartLowResolution, then 60-minute intervals thereafter), and falls
// back to today's daily precipitation total when graph data is absent.
func checkRain(plz, within int, detail *api.PLZDetail, now time.Time) rainResult {
	result := rainResult{PLZ: plz, WithinMinutes: within}

	if detail.Graph != nil && len(detail.Graph.Precipitation) > 0 {
		maxMM, ok := graphRainInWindow(detail.Graph, now, time.Duration(within)*time.Minute)
		if ok {
			result.MaxRainMM = maxMM
			result.RainExpected = maxMM > 0
			if result.RainExpected {
				result.Message = fmt.Sprintf("Rain expected: up to %.1f mm in the next %d min", maxMM, within)
			} else {
				result.Message = fmt.Sprintf("No rain expected in the next %d min", within)
			}
			return result
		}
	}

	// Fallback: today's daily total from the 10-day forecast.
	if len(detail.TenDaysForecast) > 0 {
		today := detail.TenDaysForecast[0]
		result.MaxRainMM = today.Precipitation
		result.RainExpected = today.Precipitation > 0
		if result.RainExpected {
			result.Message = fmt.Sprintf("Rain possible today: %.1f mm forecast (hourly data unavailable)", today.Precipitation)
		} else {
			result.Message = "No rain expected today (hourly data unavailable)"
		}
		return result
	}

	result.Message = "Rain data unavailable"
	return result
}

const (
	hiInterval = 10 * time.Minute // high-resolution slot width
	loInterval = 60 * time.Minute // low-resolution slot width
)

// graphRainInWindow returns the maximum precipitation value (mm) across all
// graph slots that overlap [now, now+window].
//
// The GraphData layout is:
//   - Slots [0 … hiCount-1]: 10-minute resolution, starting at Graph.Start.
//   - Slots [hiCount … N]:   60-minute resolution, starting at Graph.StartLowResolution.
//
// Returns (0, false) when now falls entirely outside the available data.
func graphRainInWindow(g *api.GraphData, now time.Time, window time.Duration) (maxMM float64, ok bool) {
	hiStart, err := parseGraphTime(g.Start)
	if err != nil {
		return 0, false
	}

	var loStart time.Time
	if g.StartLowResolution != "" {
		loStart, _ = parseGraphTime(g.StartLowResolution)
	}

	// Number of high-resolution slots before low-resolution data begins.
	hiCount := len(g.Precipitation)
	if !loStart.IsZero() {
		n := int(loStart.Sub(hiStart) / hiInterval)
		if n < hiCount {
			hiCount = n
		}
	}

	end := now.Add(window)
	for t := now; !t.After(end); {
		var idx int
		var step time.Duration

		if loStart.IsZero() || t.Before(loStart) {
			if t.Before(hiStart) {
				t = t.Add(hiInterval)
				continue
			}
			idx = int(t.Sub(hiStart) / hiInterval)
			step = hiInterval
		} else {
			idx = hiCount + int(t.Sub(loStart)/loInterval)
			step = loInterval
		}

		if idx >= len(g.Precipitation) {
			break
		}
		if idx >= 0 {
			ok = true
			if p := g.Precipitation[idx]; p > maxMM {
				maxMM = p
			}
		}
		t = t.Add(step)
	}
	return maxMM, ok
}

// parseGraphTime parses the time strings used in GraphData (Start /
// StartLowResolution), assuming Swiss local time when no zone is given.
func parseGraphTime(s string) (time.Time, error) {
	for _, f := range []string{"2006-01-02T15:04", "2006-01-02T15:04:05", time.RFC3339} {
		if t, err := time.ParseInLocation(f, s, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse graph time %q", s)
}

func printRainCheck(r rainResult) {
	icon := "☀️"
	if r.RainExpected {
		icon = "🌧️"
	}
	out.Sep(50)
	fmt.Printf("  Rain check for PLZ %d  (next %d min)\n", r.PLZ, r.WithinMinutes)
	out.Sep(50)
	fmt.Printf("  %s  %s\n", icon, r.Message)
	out.Sep(50)
}
