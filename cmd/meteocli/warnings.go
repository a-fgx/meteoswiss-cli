package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/a-fgx/meteoswiss-cli/internal/api"
	"github.com/a-fgx/meteoswiss-cli/internal/out"
)

func newWarningsCmd(flags *rootFlags) *cobra.Command {
	var plz int
	var warnLevel int

	cmd := &cobra.Command{
		Use:   "warnings",
		Short: "Show active weather warnings for a Swiss postal code",
		Example: `  # All active warnings in Bern
  meteocli warnings --zip 3000

  # Only warnings level 3 and above
  meteocli warnings --zip 3000 --min-level 3

  # Output as JSON
  meteocli warnings --zip 3000 --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if warnLevel < 1 || warnLevel > 5 {
				return fmt.Errorf("--min-level must be between 1 and 5")
			}
			if err := requirePLZ(plz); err != nil {
				return err
			}

			client := api.New()
			detail, err := client.PLZDetail(plz)
			if err != nil {
				return err
			}

			// Filter by minimum level.
			var filtered []api.Warning
			for _, w := range detail.Warnings {
				if w.WarnLevel >= warnLevel {
					filtered = append(filtered, w)
				}
			}

			if flags.asJSON {
				return out.PrintJSON(os.Stdout, filtered)
			}

			printWarnings(filtered)
			return nil
		},
	}

	cmd.Flags().IntVar(&plz, "zip", 0, "Swiss postal code (e.g. 3000 for Bern)")
	cmd.Flags().IntVar(&warnLevel, "min-level", 1, "minimum warning level to display (1=Minor … 5=Very high)")
	_ = cmd.MarkFlagRequired("zip")
	return cmd
}

func printWarnings(warnings []api.Warning) {
	if len(warnings) == 0 {
		out.Println("No active weather warnings.")
		return
	}

	out.Sep(60)
	fmt.Printf("  %d active warning(s)\n", len(warnings))
	out.Sep(60)

	for i, w := range warnings {
		wtype := api.WarnType[w.WarnType]
		if wtype == "" {
			wtype = fmt.Sprintf("Type %d", w.WarnType)
		}
		wlevel := api.WarnLevel[w.WarnLevel]
		if wlevel == "" {
			wlevel = fmt.Sprintf("Level %d", w.WarnLevel)
		}

		fmt.Printf("  [%d] %s — %s\n", i+1, wtype, wlevel)
		if w.Headline != "" {
			fmt.Printf("      %s\n", w.Headline)
		}
		if w.ValidFrom != "" || w.ValidTo != "" {
			fmt.Printf("      %s → %s\n", w.ValidFrom, w.ValidTo)
		}
		if len(w.Regions) > 0 {
			fmt.Printf("      Regions: %v\n", w.Regions)
		}
		if i < len(warnings)-1 {
			fmt.Println()
		}
	}
	out.Sep(60)
}
