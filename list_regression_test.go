package md_test

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	md "github.com/firecrawl/html-to-markdown"
)

func TestListRegression_NonLiChildInOrderedListSkipsIndex(t *testing.T) {
	conv := md.NewConverter("", true, nil)

	html := `<ol><li>one</li><div></div><li>two</li></ol>`
	out, err := conv.ConvertString(html)
	if err != nil {
		t.Fatal(err)
	}

	// goquery's Index counts element siblings, so the second <li> becomes index 2 => "3."
	if !strings.Contains(out, "1. one") {
		t.Fatalf("expected output to contain %q, got:\n%s", "1. one", out)
	}
	if !strings.Contains(out, "3. two") {
		t.Fatalf("expected output to contain %q (due to non-li sibling), got:\n%s", "3. two", out)
	}
	if strings.Contains(out, "2. two") {
		t.Fatalf("did not expect output to contain %q, got:\n%s", "2. two", out)
	}
}

func TestListRegression_WrapperListItemDoesNotEmitEmptyBullet(t *testing.T) {
	conv := md.NewConverter("", true, nil)

	html := `<ul><li><ul><li>Nested</li></ul></li></ul>`
	out, err := conv.ConvertString(html)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "Nested") {
		t.Fatalf("expected output to contain Nested, got:\n%s", out)
	}
	for _, line := range strings.Split(out, "\n") {
		if strings.TrimSpace(line) == "-" {
			t.Fatalf("unexpected empty list marker line, got:\n%s", out)
		}
	}
}

func TestListRegression_MalformedLiOutsideListDoesNotPanic(t *testing.T) {
	conv := md.NewConverter("", true, nil)

	html := `<div><li>Item</li></div>`
	out, err := conv.ConvertString(html)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Item") {
		t.Fatalf("expected output to contain Item, got:\n%s", out)
	}
}

func TestListRegression_BeforeHookMutationIsIncludedInListMetadata(t *testing.T) {
	conv := md.NewConverter("", true, nil)
	conv.Before(func(selec *goquery.Selection) {
		selec.Find("ol").First().AppendHtml("<li>b</li>")
	})

	out, err := conv.ConvertString(`<ol><li>a</li></ol>`)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "1. a") || !strings.Contains(out, "2. b") {
		t.Fatalf("expected mutated list items to be numbered, got:\n%s", out)
	}
}


