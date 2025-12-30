package md

import (
	"bytes"
	"log"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	ruleDefault = func(content string, selec *goquery.Selection, opt *Options) *string {
		return &content
	}
	ruleKeep = func(content string, selec *goquery.Selection, opt *Options) *string {
		element := selec.Get(0)

		var buf bytes.Buffer
		err := html.Render(&buf, element)
		if err != nil {
			log.Println("[firecrawl/html-to-markdown] ruleKeep: error while rendering the element to html:", err)
			return String("")
		}

		return String(buf.String())
	}
)

var inlineElements = []string{ // -> https://developer.mozilla.org/de/docs/Web/HTML/Inline_elemente
	"b", "big", "i", "small", "tt",
	"abbr", "acronym", "cite", "code", "dfn", "em", "kbd", "strong", "samp", "var",
	"a", "bdo", "br", "img", "map", "object", "q", "script", "span", "sub", "sup",
	"button", "input", "label", "select", "textarea",
	"iframe",
}

// IsInlineElement can be used to check wether a node name (goquery.Nodename) is
// an html inline element and not a block element. Used in the rule for the
// p tag to check wether the text is inside a block element.
func IsInlineElement(e string) bool {
	for _, element := range inlineElements {
		if element == e {
			return true
		}
	}
	return false
}

// String is a helper function to return a pointer.
func String(text string) *string {
	return &text
}

// Options to customize the output. You can change stuff like
// the character that is used for strong text.
type Options struct {
	// "setext" or "atx"
	// default: "atx"
	HeadingStyle string

	// Any Thematic break
	// default: "* * *"
	HorizontalRule string

	// "-", "+", or "*"
	// default: "-"
	BulletListMarker string

	// "indented" or "fenced"
	// default: "indented"
	CodeBlockStyle string

	// ``` or ~~~
	// default: ```
	Fence string

	// _ or *
	// default: _
	EmDelimiter string

	// ** or __
	// default: **
	StrongDelimiter string

	// inlined or referenced
	// default: inlined
	LinkStyle string

	// full, collapsed, or shortcut
	// default: full
	LinkReferenceStyle string

	// basic, disabled
	// default: basic
	EscapeMode string

	domain string

	// GetAbsoluteURL parses the `rawURL` and adds the `domain` to convert relative (/page.html)
	// urls to absolute urls (http://domain.com/page.html).
	//
	// The default is `DefaultGetAbsoluteURL`, unless you override it. That can also
	// be useful if you want to proxy the images.
	GetAbsoluteURL func(selec *goquery.Selection, rawURL string, domain string) string

	// GetCodeBlockLanguage identifies the language for syntax highlighting
	// of a code block. The default is `DefaultGetCodeBlockLanguage`, which
	// only gets the attribute x from the selection.
	//
	// You can override it if you want more results, for example by using
	// lexers.Analyse(content) from github.com/alecthomas/chroma
	// TODO: implement
	// GetCodeBlockLanguage func(s *goquery.Selection, content string) string
}

// DefaultGetAbsoluteURL is the default function and can be overridden through `GetAbsoluteURL` in the options.
func DefaultGetAbsoluteURL(selec *goquery.Selection, rawURL string, domain string) string {
	if domain == "" {
		return rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		// we can't do anything with this url because it is invalid
		return rawURL
	}

	if u.Scheme == "data" {
		// this is a data uri (for example an inline base64 image)
		return rawURL
	}

	if u.Scheme == "" {
		u.Scheme = "http"
	}
	if u.Host == "" {
		u.Host = domain
	}

	return u.String()
}

// AdvancedResult is used for example for links. If you use LinkStyle:referenced
// the link href is placed at the bottom of the generated markdown (Footer).
type AdvancedResult struct {
	Header   string
	Markdown string
	Footer   string
}

// Rule to convert certain html tags to markdown.
//
//	md.Rule{
//	  Filter: []string{"del", "s", "strike"},
//	  Replacement: func(content string, selec *goquery.Selection, opt *md.Options) *string {
//	    // You need to return a pointer to a string (md.String is just a helper function).
//	    // If you return nil the next function for that html element
//	    // will be picked. For example you could only convert an element
//	    // if it has a certain class name and fallback if not.
//	    return md.String("~" + content + "~")
//	  },
//	}
type Rule struct {
	Filter              []string
	Replacement         func(content string, selec *goquery.Selection, options *Options) *string
	AdvancedReplacement func(content string, selec *goquery.Selection, options *Options) (res AdvancedResult, skip bool)
}

var leadingNewlinesR = regexp.MustCompile(`^\n+`)
var trailingNewlinesR = regexp.MustCompile(`\n+$`)

var newlinesR = regexp.MustCompile(`\n+`)
var tabR = regexp.MustCompile(`\t+`)
var indentR = regexp.MustCompile(`(?m)\n`)

func (conv *Converter) applyRules(nodeName, markdown string, selec *goquery.Selection, opt *Options) (AdvancedResult, bool) {
	rules := conv.getRuleFuncs(nodeName)
	if rules == nil {
		// Tag is in remove map, return empty result
		return AdvancedResult{}, false
	}
	for i := len(rules) - 1; i >= 0; i-- {
		res, skip := rules[i](markdown, selec, opt)
		if !skip {
			return res, false
		}
	}
	return AdvancedResult{Markdown: markdown}, true
}

func appendBlock(current, addition string) string {
	if addition == "" {
		return current
	}
	if current != "" && !strings.HasSuffix(current, "\n") {
		current += "\n"
	}
	current += addition
	if !strings.HasSuffix(addition, "\n") {
		current += "\n"
	}
	return current
}

func (result *AdvancedResult) accumulate(other AdvancedResult) {
	result.Header = appendBlock(result.Header, other.Header)
	result.Footer = appendBlock(result.Footer, other.Footer)
}

func (conv *Converter) selecToMD(selec *goquery.Selection, opt *Options) AdvancedResult {
	var result AdvancedResult
	var builder strings.Builder

	selec.Contents().Each(func(i int, s *goquery.Selection) {
		nodeName := goquery.NodeName(s)

		// Check if this element should be removed before processing
		conv.mutex.RLock()
		_, shouldRemove := conv.remove[nodeName]
		conv.mutex.RUnlock()

		if shouldRemove {
			// Skip processing removed elements entirely
			return
		}

		content := conv.selecToMD(s, opt)
		result.accumulate(content)

		ruleResult, useOriginal := conv.applyRules(nodeName, content.Markdown, s, opt)
		result.accumulate(ruleResult)

		if !useOriginal {
			builder.WriteString(ruleResult.Markdown)
		} else {
			builder.WriteString(content.Markdown)
		}
	})

	result.Markdown = builder.String()
	return result
}

func (conv *Converter) applyRulesToSelection(selec *goquery.Selection, opt *Options) AdvancedResult {
	if selec == nil || len(selec.Nodes) == 0 {
		return AdvancedResult{}
	}

	content := conv.selecToMD(selec, opt)
	result := AdvancedResult{
		Header: content.Header,
		Footer: content.Footer,
	}

	ruleResult, useOriginal := conv.applyRules(goquery.NodeName(selec), content.Markdown, selec, opt)
	result.accumulate(ruleResult)

	if !useOriginal {
		result.Markdown = ruleResult.Markdown
	} else {
		result.Markdown = content.Markdown
	}

	return result
}
