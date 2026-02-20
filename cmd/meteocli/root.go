package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/meteocli/internal/out"
)

var version = "0.1.0"

type rootFlags struct {
	asJSON bool
}

func execute(args []string) error {
	var flags rootFlags

	rootCmd := &cobra.Command{
		Use:   "meteocli",
		Short: "meteocli — weather data from MeteoSwiss, right in your terminal",
		Long: `meteocli is a command-line interface for MeteoSwiss, the Swiss federal
meteorological service. It fetches current conditions, multi-day forecasts,
and active weather warnings from the MeteoSwiss app backend.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}
	rootCmd.SetVersionTemplate("meteocli {{.Version}}\n")

	rootCmd.PersistentFlags().BoolVar(&flags.asJSON, "json", false, "output JSON instead of human-readable text")

	rootCmd.AddCommand(newVersionCmd())
	rootCmd.AddCommand(newWeatherCmd(&flags))
	rootCmd.AddCommand(newForecastCmd(&flags))
	rootCmd.AddCommand(newWarningsCmd(&flags))

	rootCmd.SetArgs(args)
	if err := rootCmd.Execute(); err != nil {
		_ = out.WriteError(os.Stderr, flags.asJSON, err)
		return err
	}
	return nil
}

// requirePLZ validates that a postal code looks like a valid Swiss PLZ.
func requirePLZ(plz int) error {
	if plz < 1000 || plz > 9999 {
		return fmt.Errorf("invalid Swiss postal code %d: must be between 1000 and 9999", plz)
	}
	return nil
}
