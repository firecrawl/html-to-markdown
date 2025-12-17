package md_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	md "github.com/firecrawl/html-to-markdown"
)

// TestPerfBigHTML_Smoke runs a large HTML file (random-ish content),
// converts it once, and asserts we get non-empty output.
//
// This is intentionally a "smoke" perf test: no golden output, just "it renders".
// Run it alone to track timings over time:
//
//	go test -run '^TestPerfBigHTML_Smoke$' -count=1
func TestPerfBigHTML_Smoke(t *testing.T) {
	p := filepath.Join("testdata", "Perf", "big.html")

	if _, err := os.Stat(p); err != nil {
		p = filepath.Join("testdata", "Perf", "test.html")
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read %s: %v", p, err)
	}
	html := string(b)

	conv := md.NewConverter("", true, nil)
	out, err := conv.ConvertString(html)
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	if strings.TrimSpace(out) == "" {
		t.Fatalf("expected non-empty markdown output")
	}
}
