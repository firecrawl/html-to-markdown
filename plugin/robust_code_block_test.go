package plugin

import (
	"strings"
	"testing"

	md "github.com/firecrawl/html-to-markdown"
)

func TestRobustCodeBlock_LanguageDetection(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "language- prefix on code",
			html:     `<pre><code class="language-javascript">const x = 1;</code></pre>`,
			expected: "```javascript",
		},
		{
			name:     "lang- prefix on code",
			html:     `<pre><code class="lang-python">print("hello")</code></pre>`,
			expected: "```python",
		},
		{
			name:     "language on pre element",
			html:     `<pre class="language-go"><code>func main() {}</code></pre>`,
			expected: "```go",
		},
		{
			name:     "no language class",
			html:     `<pre><code>plain code</code></pre>`,
			expected: "```\nplain code",
		},
	}

	conv := md.NewConverter("", true, nil)
	conv.Use(RobustCodeBlock())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markdown, err := conv.ConvertString(tt.html)
			if err != nil {
				t.Fatalf("ConvertString failed: %v", err)
			}

			if !strings.Contains(markdown, tt.expected) {
				t.Errorf("expected output to contain %q\nGot:\n%s", tt.expected, markdown)
			}
		})
	}
}

func TestRobustCodeBlock_GutterStripping(t *testing.T) {
	tests := []struct {
		name        string
		html        string
		shouldHave  []string
		shouldntHave []string
	}{
		{
			name: "strips gutter class",
			html: `<pre><code><table><tr><td class="gutter">1
2
3</td><td class="code">const a = 1;
const b = 2;
const c = 3;</td></tr></table></code></pre>`,
			shouldHave:   []string{"const a = 1", "const b = 2", "const c = 3"},
			shouldntHave: []string{"\n1\n2\n3"},
		},
		{
			name: "strips line-numbers class",
			html: `<pre><code><div class="line-numbers">1
2</div><div class="content">hello
world</div></code></pre>`,
			shouldHave:   []string{"hello", "world"},
			shouldntHave: []string{},
		},
	}

	conv := md.NewConverter("", true, nil)
	conv.Use(RobustCodeBlock())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markdown, err := conv.ConvertString(tt.html)
			if err != nil {
				t.Fatalf("ConvertString failed: %v", err)
			}

			for _, exp := range tt.shouldHave {
				if !strings.Contains(markdown, exp) {
					t.Errorf("expected output to contain %q\nGot:\n%s", exp, markdown)
				}
			}
			for _, notExp := range tt.shouldntHave {
				if strings.Contains(markdown, notExp) {
					t.Errorf("expected output NOT to contain %q\nGot:\n%s", notExp, markdown)
				}
			}
		})
	}
}

func TestRobustCodeBlock_SyntaxHighlighterDivs(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected []string
	}{
		{
			name: "highlight.js style",
			html: `<pre><code class="language-json"><div class="hljs">{
  "key": "value"
}</div></code></pre>`,
			expected: []string{`"key"`, `"value"`},
		},
		{
			name: "prism.js style with spans",
			html: `<pre><code class="language-javascript"><span class="token keyword">const</span> <span class="token variable">x</span> <span class="token operator">=</span> <span class="token number">42</span><span class="token punctuation">;</span></code></pre>`,
			expected: []string{"const", "x", "=", "42"},
		},
		{
			name: "nested divs with token-line",
			html: `<pre><code><div class="token-line">line1</div><div class="token-line">line2</div></code></pre>`,
			expected: []string{"line1", "line2"},
		},
	}

	conv := md.NewConverter("", true, nil)
	conv.Use(RobustCodeBlock())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markdown, err := conv.ConvertString(tt.html)
			if err != nil {
				t.Fatalf("ConvertString failed: %v", err)
			}

			for _, exp := range tt.expected {
				if !strings.Contains(markdown, exp) {
					t.Errorf("expected output to contain %q\nGot:\n%s", exp, markdown)
				}
			}
		})
	}
}

func TestRobustCodeBlock_InlineCode(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "simple inline code",
			html:     `<p>Use <code>fmt.Println</code> to print</p>`,
			expected: "`fmt.Println`",
		},
		{
			name:     "inline code with backticks",
			html:     `<p>Use <code>echo ` + "`hello`" + `</code> command</p>`,
			expected: "``echo `hello```",
		},
		{
			name:     "inline code not affected by pre rule",
			html:     `<p>The <code>main</code> function is important</p>`,
			expected: "`main`",
		},
	}

	conv := md.NewConverter("", true, nil)
	conv.Use(RobustCodeBlock())

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			markdown, err := conv.ConvertString(tt.html)
			if err != nil {
				t.Fatalf("ConvertString failed: %v", err)
			}

			if !strings.Contains(markdown, tt.expected) {
				t.Errorf("expected output to contain %q\nGot:\n%s", tt.expected, markdown)
			}
		})
	}
}

func TestRobustCodeBlock_PreservesNewlines(t *testing.T) {
	html := `<pre><code class="language-python">def hello():
    print("world")

hello()</code></pre>`

	conv := md.NewConverter("", true, nil)
	conv.Use(RobustCodeBlock())

	markdown, err := conv.ConvertString(html)
	if err != nil {
		t.Fatalf("ConvertString failed: %v", err)
	}

	// Check structure is preserved
	if !strings.Contains(markdown, "def hello():") {
		t.Error("missing function definition")
	}
	if !strings.Contains(markdown, `print("world")`) {
		t.Error("missing print statement")
	}
	if !strings.Contains(markdown, "hello()") {
		t.Error("missing function call")
	}
}

func TestRobustCodeBlock_BrTags(t *testing.T) {
	html := `<pre><code>line1<br>line2<br/>line3</code></pre>`

	conv := md.NewConverter("", true, nil)
	conv.Use(RobustCodeBlock())

	markdown, err := conv.ConvertString(html)
	if err != nil {
		t.Fatalf("ConvertString failed: %v", err)
	}

	if !strings.Contains(markdown, "line1") {
		t.Error("missing line1")
	}
	if !strings.Contains(markdown, "line2") {
		t.Error("missing line2")
	}
	if !strings.Contains(markdown, "line3") {
		t.Error("missing line3")
	}
}
