package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/meteocli/internal/api"
	"github.com/user/meteocli/internal/out"
)

func newForecastCmd(flags *rootFlags) *cobra.Command {
	var plz int
	var days int

	cmd := &cobra.Command{
		Use:   "forecast",
		Short: "Show the multi-day weather forecast for a Swiss postal code",
		Example: `  # 7-day forecast for Zurich
  meteocli forecast --zip 8000

  # 3-day forecast for Geneva as JSON
  meteocli forecast --zip 1200 --days 3 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePLZ(plz); err != nil {
				return err
			}
			if days < 1 || days > 10 {
				return fmt.Errorf("--days must be between 1 and 10")
			}

			client := api.New()
			detail, err := client.PLZDetail(plz)
			if err != nil {
				return err
			}

			forecast := detail.Forecast
			if len(forecast) > days {
				forecast = forecast[:days]
			}

			if flags.asJSON {
				return out.PrintJSON(os.Stdout, forecast)
			}

			printForecast(plz, forecast)
			return nil
		},
	}

	cmd.Flags().IntVar(&plz, "zip", 0, "Swiss postal code (e.g. 8000 for Zurich)")
	cmd.Flags().IntVar(&days, "days", 7, "number of days to show (1–10)")
	_ = cmd.MarkFlagRequired("zip")
	return cmd
}

func printForecast(plz int, forecast []api.DayForecast) {
	out.Sep(60)
	fmt.Printf("  %d-day forecast for PLZ %d\n", len(forecast), plz)
	out.Sep(60)
	fmt.Printf("  %-12s %-22s %6s %6s  %8s\n", "Date", "Conditions", "Min°C", "Max°C", "Rain mm")
	out.Sep(60)

	for _, day := range forecast {
		emoji := api.IconEmoji(day.IconDay)
		desc := api.IconDescription(day.IconDay)
		label := fmt.Sprintf("%s (%s)", desc, emoji)
		fmt.Printf("  %-12s %-22s %6.1f %6.1f  %8.1f\n",
			day.DayDate,
			truncate(label, 22),
			day.TemperatureMin,
			day.TemperatureMax,
			day.Precipitation,
		)
	}
	out.Sep(60)
}

// truncate shortens s to at most n runes.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
