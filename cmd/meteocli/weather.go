package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/meteocli/internal/api"
	"github.com/user/meteocli/internal/out"
)

func newWeatherCmd(flags *rootFlags) *cobra.Command {
	var plz int

	cmd := &cobra.Command{
		Use:   "weather",
		Short: "Show current weather conditions for a Swiss postal code",
		Example: `  # Current weather in Zurich
  meteocli weather --zip 8000

  # Current weather in Bern as JSON
  meteocli weather --zip 3000 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePLZ(plz); err != nil {
				return err
			}

			client := api.New()
			detail, err := client.PLZDetail(plz)
			if err != nil {
				return err
			}

			if flags.asJSON {
				return out.PrintJSON(os.Stdout, detail.CurrentWeather)
			}

			printCurrentWeather(plz, detail)
			return nil
		},
	}

	cmd.Flags().IntVar(&plz, "zip", 0, "Swiss postal code (e.g. 8000 for Zurich)")
	_ = cmd.MarkFlagRequired("zip")
	return cmd
}

func printCurrentWeather(plz int, detail *api.PLZDetail) {
	cw := detail.CurrentWeather
	emoji := api.IconEmoji(cw.Icon)
	desc := api.IconDescription(cw.Icon)

	out.Sep(44)
	fmt.Printf("  Weather for PLZ %d\n", plz)
	out.Sep(44)
	fmt.Printf("  %s (%s)\n", desc, emoji)
	fmt.Printf("  Temperature : %.1f °C\n", cw.Temperature)
	if cw.Time != "" {
		fmt.Printf("  Observed at : %s\n", cw.Time)
	}
	out.Sep(44)

	// Show today's forecast summary if available.
	if len(detail.TenDaysForecast) > 0 {
		today := detail.TenDaysForecast[0]
		fmt.Printf("  Today       : %.1f / %.1f °C  rain %.1f mm\n",
			today.TemperatureMin, today.TemperatureMax, today.Precipitation)
		if today.WindSpeed > 0 {
			fmt.Printf("  Wind        : %s  %d km/h\n",
				api.WindDirectionLabel(today.WindDirection), today.WindSpeed)
		}
		out.Sep(44)
	}
}
