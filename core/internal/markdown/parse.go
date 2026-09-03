package markdown

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/text"
)

var md = goldmark.New(goldmark.WithExtensions(extension.GFM))

func Parse(source []byte) ([]Block, error) {
	reader := text.NewReader(source)
	doc := md.Parser().Parse(reader)

	var blocks []Block

	err := ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Heading:
			bt := Heading3
			switch node.Level {
			case 1:
				bt = Heading1
			case 2:
				bt = Heading2
			}
			blocks = append(blocks, Block{Type: bt, Text: inlineText(node, source)})
			return ast.WalkSkipChildren, nil

		case *ast.Paragraph:
			if _, inList := node.Parent().(*ast.ListItem); inList {
				return ast.WalkContinue, nil
			}
			blocks = append(blocks, Block{Type: Paragraph, Text: inlineText(node, source)})
			return ast.WalkSkipChildren, nil

		case *ast.ListItem:
			blocks = append(blocks, listItemBlock(node, source))
			return ast.WalkSkipChildren, nil

		case *ast.Blockquote:
			blocks = append(blocks, Block{Type: Quote, Text: inlineText(node, source)})
			return ast.WalkSkipChildren, nil

		case *ast.FencedCodeBlock:
			blocks = append(blocks, codeBlock(node, source))
			return ast.WalkSkipChildren, nil

		case *ast.ThematicBreak:
			blocks = append(blocks, Block{Type: Divider})
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})

	return blocks, err
}

func listItemBlock(item *ast.ListItem, source []byte) Block {
	if tc := findTaskCheckbox(item); tc != nil {
		b := Block{Type: ToDoItem, Checked: tc.IsChecked, Text: inlineText(item, source)}
		return b
	}
	list, _ := item.Parent().(*ast.List)
	if list != nil && list.IsOrdered() {
		return Block{Type: NumberItem, Text: inlineText(item, source)}
	}
	return Block{Type: BulletItem, Text: inlineText(item, source)}
}

func findTaskCheckbox(n ast.Node) *extast.TaskCheckBox {
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		for gc := c.FirstChild(); gc != nil; gc = gc.NextSibling() {
			if tc, ok := gc.(*extast.TaskCheckBox); ok {
				return tc
			}
		}
	}
	return nil
}

func codeBlock(node *ast.FencedCodeBlock, source []byte) Block {
	var content string
	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		content += string(line.Value(source))
	}
	lang := string(node.Language(source))
	if lang == "" {
		lang = "plain text"
	}
	return Block{
		Type:     Code,
		Language: lang,
		Text:     []RichText{{Content: content}},
	}
}
