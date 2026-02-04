package plugin

import (
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
	md "github.com/firecrawl/html-to-markdown"
	"golang.org/x/net/html"
)

// RobustCodeBlock adds a robust PRE/CODE handler that extracts nested code text
// (e.g., from syntax highlighters with tables/rows/gutters) and outputs fenced
// blocks with detected language. This is useful for scraping code blocks from
// various websites that use different syntax highlighting libraries.
func RobustCodeBlock() md.Plugin {
	return func(c *md.Converter) []md.Rule {
		isGutter := func(class string) bool {
			lower := strings.ToLower(class)
			return strings.Contains(lower, "gutter") || strings.Contains(lower, "line-numbers")
		}

		detectLang := func(sel *goquery.Selection) string {
			classes := sel.AttrOr("class", "")
			lower := strings.ToLower(classes)
			for _, part := range strings.Fields(lower) {
				if strings.HasPrefix(part, "language-") {
					return strings.TrimPrefix(part, "language-")
				}
				if strings.HasPrefix(part, "lang-") {
					return strings.TrimPrefix(part, "lang-")
				}
			}
			return ""
		}

		// collect extracts text recursively, inserting newlines after block elements and br
		var collect func(n *html.Node, b *strings.Builder)
		collect = func(n *html.Node, b *strings.Builder) {
			if n == nil {
				return
			}
			switch n.Type {
			case html.TextNode:
				b.WriteString(n.Data)
			case html.ElementNode:
				name := strings.ToLower(n.Data)
				// Skip gutters
				if name != "" {
					for _, a := range n.Attr {
						if a.Key == "class" && isGutter(a.Val) {
							return
						}
					}
				}

				if name == "br" {
					b.WriteString("\n")
				}

				for child := n.FirstChild; child != nil; child = child.NextSibling {
					collect(child, b)
				}

				// Newline after block-ish wrappers to preserve lines
				switch name {
				case "p", "div", "li", "tr", "table", "thead", "tbody", "tfoot", "section", "article", "blockquote", "pre", "h1", "h2", "h3", "h4", "h5", "h6":
					b.WriteString("\n")
				}
			}
		}

		preRule := md.Rule{
			Filter: []string{"pre"},
			Replacement: func(_ string, selec *goquery.Selection, opt *md.Options) *string {
				// Find inner <code> if present for language detection
				codeSel := selec.Find("code").First()
				lang := detectLang(codeSel)
				if lang == "" {
					lang = detectLang(selec)
				}

				var b strings.Builder
				for _, n := range selec.Nodes {
					collect(n, &b)
				}
				content := strings.TrimRight(b.String(), "\n")

				fenceChar, _ := utf8.DecodeRuneInString(opt.Fence)
				fence := md.CalculateCodeFence(fenceChar, content)
				text := "\n\n" + fence + lang + "\n" + content + "\n" + fence + "\n\n"
				return md.String(text)
			},
		}

		codeRule := md.Rule{
			Filter: []string{"code"},
			Replacement: func(_ string, selec *goquery.Selection, opt *md.Options) *string {
				// If inside pre, let the PRE rule handle it
				if selec.ParentsFiltered("pre").Length() > 0 {
					return nil
				}

				var b strings.Builder
				for _, n := range selec.Nodes {
					collect(n, &b)
				}
				code := b.String()
				// Collapse multiple newlines for inline code
				code = md.TrimTrailingSpaces(strings.ReplaceAll(code, "\r\n", "\n"))

				// Choose fence length safely
				fence := "`"
				if strings.Contains(code, "`") {
					fence = "``"
					if strings.Contains(code, "``") {
						fence = "```"
					}
				}
				out := fence + code + fence
				return md.String(out)
			},
		}

		return []md.Rule{preRule, codeRule}
	}
}
