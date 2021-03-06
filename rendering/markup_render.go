package rendering

import (
	"html"
	"regexp"

	"github.com/microcosm-cc/bluemonday"
)

// IsMarkupSupported indicates if the given markup is supported
func IsMarkupSupported(markup string) bool {
	if markup == SystemMarkupDefault || markup == SystemMarkupMarkdown {
		return true
	}
	return false
}

// RenderMarkupToHTML converts the given `content` in HTML using the markup tool corresponding to the given `markup` argument
// or return nil if no tool for the given `markup` is available, or returns an `error` if the command was not found or failed.
func RenderMarkupToHTML(content, markup string) string {
	switch markup {
	case SystemMarkupPlainText:
		return html.EscapeString(content)
	case SystemMarkupMarkdown:
		unsafe := MarkdownCommonHighlighter([]byte(content))
		p := bluemonday.UGCPolicy()
		p.AllowAttrs("class").Matching(regexp.MustCompile("^language-[a-zA-Z0-9]+$|prettyprint")).OnElements("code")
		p.AllowAttrs("class").OnElements("span")
		p.AllowElements("input")
		p.AllowAttrs("type").OnElements("input")
		p.AllowAttrs("checked").OnElements("input")
		p.AllowAttrs("disabled").OnElements("input")
		p.AllowAttrs("data-checkbox-index").OnElements("input")
		p.AllowAttrs("class").OnElements("input")
		html := string(p.SanitizeBytes(unsafe))
		return html
	default:
		return ""
	}
}
