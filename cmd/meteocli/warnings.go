package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/meteocli/internal/api"
	"github.com/user/meteocli/internal/out"
)

func newWarningsCmd(flags *rootFlags) *cobra.Command {
	var warnLevel int

	cmd := &cobra.Command{
		Use:   "warnings",
		Short: "Show active MeteoSwiss weather warnings for Switzerland",
		Example: `  # All active warnings
  meteocli warnings

  # Only warnings level 3 and above
  meteocli warnings --min-level 3

  # Output as JSON
  meteocli warnings --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if warnLevel < 1 || warnLevel > 5 {
				return fmt.Errorf("--min-level must be between 1 and 5")
			}

			client := api.New()
			warnings, err := client.Warnings()
			if err != nil {
				return err
			}

			// Filter by minimum level.
			filtered := warnings[:0]
			for _, w := range warnings {
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

	cmd.Flags().IntVar(&warnLevel, "min-level", 1, "minimum warning level to display (1=Minor … 5=Very high)")
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
