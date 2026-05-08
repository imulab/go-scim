// Spike candidate B: text/template + go/format.Source.
//
// This runner exists to expose the conditional-import bookkeeping cost that
// templates pay when the emitted artifact's imports vary by IR shape. The
// counterpart in spike/jennifer/main.go produces the same output via the
// dave/jennifer programmatic AST. See spike/README.md and PROJECT.md
// "Key Decisions" for the comparison.
package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"text/template"
)

// userData holds the spike fixture toggles. The runner hard-codes a fixture
// that turns on every conditional import so the gnarly path is exercised.
type userData struct {
	HasTime    bool
	HasRegexp  bool
	HasErrors  bool
	HasSlices  bool
}

func main() {
	tmplBytes, err := os.ReadFile("user.go.tmpl")
	if err != nil {
		fmt.Fprintf(os.Stderr, "read template: %v\n", err)
		os.Exit(1)
	}

	tmpl := template.Must(template.New("user").Parse(string(tmplBytes)))

	fixture := userData{
		HasTime:   true,
		HasRegexp: true,
		HasErrors: true,
		HasSlices: true,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, fixture); err != nil {
		fmt.Fprintf(os.Stderr, "execute template: %v\n", err)
		os.Exit(1)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		// Per 00-RESEARCH.md Pitfall 9: fail loud with the offending raw output
		// so the template author can see what produced invalid Go.
		fmt.Fprintf(os.Stderr, "format.Source: %v\nraw output:\n%s\n", err, buf.String())
		os.Exit(1)
	}

	if err := os.MkdirAll("output", 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir output: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile("output/user_gen.go", formatted, 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "write output: %v\n", err)
		os.Exit(1)
	}
}
