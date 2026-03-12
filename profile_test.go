package md_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	md "github.com/firecrawl/html-to-markdown"
	"github.com/firecrawl/html-to-markdown/plugin"
)

func loadTestFile(b *testing.B, name string) string {
	b.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "Perf", name))
	if err != nil {
		b.Fatalf("read %s: %v", name, err)
	}
	return string(data)
}

// generateBlockquoteHTML creates a blockquote-heavy HTML document.
func generateBlockquoteHTML(count int) string {
	var sb strings.Builder
	sb.WriteString("<html><body>")
	for i := 0; i < count; i++ {
		sb.WriteString(fmt.Sprintf(`<blockquote><p>Quote number %d with some text content that is long enough to be realistic.</p></blockquote>`, i))
		if i%5 == 0 {
			// nested blockquote every 5th
			sb.WriteString(fmt.Sprintf(`<blockquote><p>Outer %d</p><blockquote><p>Nested inner blockquote %d</p></blockquote></blockquote>`, i, i))
		}
	}
	sb.WriteString("</body></html>")
	return sb.String()
}

// generateMixedHTML creates HTML with a mix of all element types.
func generateMixedHTML(sections int) string {
	var sb strings.Builder
	sb.WriteString("<html><body>")
	for i := 0; i < sections; i++ {
		// headings + paragraphs + text nodes (escape path)
		sb.WriteString(fmt.Sprintf(`<h2>Section %d</h2>`, i))
		sb.WriteString(fmt.Sprintf(`<p>This is paragraph %d with <strong>bold</strong> and <em>italic</em> text. It has * special + markdown - characters.</p>`, i))

		// blockquote
		sb.WriteString(fmt.Sprintf(`<blockquote><p>A quote in section %d</p></blockquote>`, i))

		// list
		sb.WriteString(`<ul>`)
		for j := 0; j < 5; j++ {
			sb.WriteString(fmt.Sprintf(`<li>Item %d-%d with <a href="https://example.com/%d">a link</a></li>`, i, j, j))
		}
		sb.WriteString(`</ul>`)

		// table
		sb.WriteString(`<table><thead><tr>`)
		for j := 0; j < 4; j++ {
			sb.WriteString(fmt.Sprintf(`<th>Col %d</th>`, j))
		}
		sb.WriteString(`</tr></thead><tbody>`)
		for j := 0; j < 3; j++ {
			sb.WriteString(`<tr>`)
			for k := 0; k < 4; k++ {
				sb.WriteString(fmt.Sprintf(`<td>Cell %d-%d-%d</td>`, i, j, k))
			}
			sb.WriteString(`</tr>`)
		}
		sb.WriteString(`</tbody></table>`)

		// code block
		sb.WriteString(fmt.Sprintf(`<pre><code>func example%d() { return }</code></pre>`, i))
	}
	sb.WriteString("</body></html>")
	return sb.String()
}

// generateEscapeHeavyHTML creates HTML with lots of text nodes containing
// markdown-special characters, which exercises the escape.MarkdownCharacters path.
func generateEscapeHeavyHTML(count int) string {
	var sb strings.Builder
	sb.WriteString("<html><body>")
	for i := 0; i < count; i++ {
		sb.WriteString(fmt.Sprintf(`<p>Line %d: Price is $10.00 * quantity + tax - discount. Use [brackets] and > arrows. List: * item1 + item2 - item3. Code: `+"`"+`inline`+"`"+`.</p>`, i))
	}
	sb.WriteString("</body></html>")
	return sb.String()
}

func BenchmarkBigHTML(b *testing.B) {
	html := loadTestFile(b, "big.html")
	conv := md.NewConverter("", true, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.ConvertString(html)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBigList(b *testing.B) {
	html := loadTestFile(b, "big_list.html")
	conv := md.NewConverter("", true, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.ConvertString(html)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBigTable(b *testing.B) {
	html := loadTestFile(b, "table.html")
	conv := md.NewConverter("", true, nil)
	conv.Use(plugin.GitHubFlavored())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.ConvertString(html)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBlockquoteHeavy(b *testing.B) {
	html := generateBlockquoteHTML(500)
	conv := md.NewConverter("", true, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.ConvertString(html)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEscapeHeavy(b *testing.B) {
	html := generateEscapeHeavyHTML(1000)
	conv := md.NewConverter("", true, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.ConvertString(html)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMixed(b *testing.B) {
	html := generateMixedHTML(200)
	conv := md.NewConverter("", true, nil)
	conv.Use(plugin.GitHubFlavored())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.ConvertString(html)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRealWorld_BlogGolang(b *testing.B) {
	data, err := os.ReadFile(filepath.Join("testdata", "TestRealWorld", "blog.golang.org", "input.html"))
	if err != nil {
		b.Fatal(err)
	}
	html := string(data)
	conv := md.NewConverter("", true, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := conv.ConvertString(html)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrent simulates the service under load: multiple goroutines
// sharing a single Converter (same as the go-html-to-md-service does).
func BenchmarkConcurrent(b *testing.B) {
	html := generateMixedHTML(50)
	conv := md.NewConverter("", true, nil)
	conv.Use(plugin.GitHubFlavored())
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := conv.ConvertString(html)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
