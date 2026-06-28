package md

import (
	"testing"
)

// Regression test for https://github.com/firecrawl/firecrawl/issues/3583
//
// Whitespace between adjacent inline elements is significant in HTML: it
// collapses to a single space rather than being removed. Previously the
// converter dropped whitespace-only text nodes entirely, gluing adjacent
// inline elements together (e.g. cookie-consent banners produced
// "FunktionalFunktional"). Whitespace at a block boundary must still be
// dropped so block separation (\n\n) is unaffected.
func TestInlineWhitespacePreserved(t *testing.T) {
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			name: "spans separated by a space",
			html: `<span>a</span> <span>b</span>`,
			want: "a b",
		},
		{
			name: "spans separated by a newline",
			html: "<span>a</span>\n<span>b</span>",
			want: "a b",
		},
		{
			name: "labels separated by a newline and indent",
			html: "<label>a</label>\n    <label>b</label>",
			want: "a b",
		},
		{
			name: "no whitespace in source stays glued (matches rendering)",
			html: `<span>a</span><span>b</span>`,
			want: "ab",
		},
		{
			name: "cookie-consent banner (issue #3583 repro)",
			html: "<div>\n" +
				`  <input type="checkbox" id="x">` + "\n" +
				`  <label for="x">Funktional</label>` + "\n" +
				`  <span>Funktional</span>` + "\n" +
				"</div>\n<div>Always active</div>",
			want: "Funktional Funktional\n\nAlways active",
		},
		{
			name: "block boundary whitespace is not turned into a space",
			html: `<div>a</div> <div>b</div>`,
			want: "a\n\nb",
		},
		{
			name: "leading/trailing whitespace inside a block is dropped",
			html: "<div>\n  <span>a</span>\n</div>",
			want: "a",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv := NewConverter("", true, nil)
			got, err := conv.ConvertString(tt.html)
			if err != nil {
				t.Fatalf("ConvertString error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ConvertString(%q)\n got: %q\nwant: %q", tt.html, got, tt.want)
			}
		})
	}
}
