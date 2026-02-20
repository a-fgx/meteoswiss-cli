// Package out provides helpers for writing human-readable or JSON output to
// an io.Writer.
package out

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Print writes a human-readable message to stdout.
func Print(format string, args ...any) {
	fmt.Fprintf(os.Stdout, format, args...)
}

// Println writes a human-readable line to stdout.
func Println(s string) {
	fmt.Fprintln(os.Stdout, s)
}

// PrintJSON marshals v to indented JSON and writes it to w.
func PrintJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

// WriteError writes err to w; if asJSON is true it uses a JSON envelope.
func WriteError(w io.Writer, asJSON bool, err error) error {
	if asJSON {
		return PrintJSON(w, map[string]string{"error": err.Error()})
	}
	fmt.Fprintf(w, "Error: %v\n", err)
	return nil
}

// Sep prints a separator line of n dashes to stdout.
func Sep(n int) {
	for i := 0; i < n; i++ {
		fmt.Fprint(os.Stdout, "─")
	}
	fmt.Fprintln(os.Stdout)
}
