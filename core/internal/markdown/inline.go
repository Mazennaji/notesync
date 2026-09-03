package markdown

import (
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

func inlineText(n ast.Node, source []byte) []RichText {
	var spans []RichText
	collectInline(n, source, RichText{}, &spans)
	if len(spans) == 0 {
		spans = append(spans, RichText{Content: ""})
	}
	return spans
}

func collectInline(n ast.Node, source []byte, style RichText, out *[]RichText) {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch node := c.(type) {
		case *ast.Text:
			s := style
			s.Content = string(node.Segment.Value(source))
			if s.Content != "" {
				*out = append(*out, s)
			}
		case *ast.String:
			s := style
			s.Content = string(node.Value)
			*out = append(*out, s)

		case *ast.Emphasis:
			next := style
			if node.Level == 2 {
				next.Bold = true
			} else {
				next.Italic = true
			}
			collectInline(node, source, next, out)

		case *ast.CodeSpan:
			s := style
			s.Code = true
			s.Content = string(node.Text(source))
			*out = append(*out, s)

		case *extast.Strikethrough:
			next := style
			next.Strike = true
			collectInline(node, source, next, out)

		case *ast.Link:
			next := style
			next.Link = string(node.Destination)
			collectInline(node, source, next, out)

		default:
			collectInline(c, source, style, out)
		}
	}
}
