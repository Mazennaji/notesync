package notion

import "github.com/Mazennaji/notesync/core/internal/markdown"

const maxTextLen = 2000

func buildBlocks(blocks []markdown.Block) []map[string]any {
	out := make([]map[string]any, 0, len(blocks))
	for _, b := range blocks {
		out = append(out, buildBlock(b))
	}
	return out
}

func buildBlock(b markdown.Block) map[string]any {
	obj := map[string]any{
		"object": "block",
		"type":   string(b.Type),
	}

	switch b.Type {
	case markdown.Divider:
		obj[string(b.Type)] = map[string]any{}

	case markdown.Code:
		obj[string(b.Type)] = map[string]any{
			"rich_text": buildRichText(b.Text),
			"language":  b.Language,
		}

	case markdown.ToDoItem:
		obj[string(b.Type)] = map[string]any{
			"rich_text": buildRichText(b.Text),
			"checked":   b.Checked,
		}

	default:
		obj[string(b.Type)] = map[string]any{
			"rich_text": buildRichText(b.Text),
		}
	}
	return obj
}

func buildRichText(spans []markdown.RichText) []map[string]any {
	out := make([]map[string]any, 0, len(spans))
	for _, s := range spans {
		content := s.Content
		if len(content) > maxTextLen {
			content = content[:maxTextLen]
		}
		text := map[string]any{"content": content}
		if s.Link != "" {
			text["link"] = map[string]any{"url": s.Link}
		}
		out = append(out, map[string]any{
			"type": "text",
			"text": text,
			"annotations": map[string]any{
				"bold":          s.Bold,
				"italic":        s.Italic,
				"strikethrough": s.Strike,
				"code":          s.Code,
			},
		})
	}
	return out
}
