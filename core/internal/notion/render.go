package notion

import (
	"encoding/json"
	"strings"
)

func (c *Client) FetchMarkdown(pageID string) (string, error) {
	blocks, err := c.fetchChildren(pageID)
	if err != nil {
		return "", err
	}
	return renderBlocks(blocks), nil
}

type notionBlock struct {
	Type      string         `json:"type"`
	Paragraph *richContainer `json:"paragraph,omitempty"`
	Heading1  *richContainer `json:"heading_1,omitempty"`
	Heading2  *richContainer `json:"heading_2,omitempty"`
	Heading3  *richContainer `json:"heading_3,omitempty"`
	Bulleted  *richContainer `json:"bulleted_list_item,omitempty"`
	Numbered  *richContainer `json:"numbered_list_item,omitempty"`
	ToDo      *todoContainer `json:"to_do,omitempty"`
	Quote     *richContainer `json:"quote,omitempty"`
	Code      *codeContainer `json:"code,omitempty"`
}

type richContainer struct {
	RichText []notionRichText `json:"rich_text"`
}

type todoContainer struct {
	RichText []notionRichText `json:"rich_text"`
	Checked  bool             `json:"checked"`
}

type codeContainer struct {
	RichText []notionRichText `json:"rich_text"`
	Language string           `json:"language"`
}

type notionRichText struct {
	PlainText   string `json:"plain_text"`
	Href        string `json:"href"`
	Annotations struct {
		Bold          bool `json:"bold"`
		Italic        bool `json:"italic"`
		Strikethrough bool `json:"strikethrough"`
		Code          bool `json:"code"`
	} `json:"annotations"`
}

func (c *Client) fetchChildren(pageID string) ([]notionBlock, error) {
	var blocks []notionBlock
	cursor := ""
	for {
		path := "/blocks/" + pageID + "/children?page_size=100"
		if cursor != "" {
			path += "&start_cursor=" + cursor
		}
		data, err := c.do("GET", path, nil)
		if err != nil {
			return nil, err
		}
		var resp struct {
			Results    []notionBlock `json:"results"`
			NextCursor string        `json:"next_cursor"`
			HasMore    bool          `json:"has_more"`
		}
		if err := json.Unmarshal(data, &resp); err != nil {
			return nil, err
		}
		blocks = append(blocks, resp.Results...)
		if !resp.HasMore || resp.NextCursor == "" {
			break
		}
		cursor = resp.NextCursor
	}
	return blocks, nil
}

func renderBlocks(blocks []notionBlock) string {
	var sb strings.Builder
	numberCounter := 0

	for _, b := range blocks {
		if b.Type != "numbered_list_item" {
			numberCounter = 0
		}
		switch b.Type {
		case "heading_1":
			sb.WriteString("# " + renderRich(b.Heading1.RichText) + "\n\n")
		case "heading_2":
			sb.WriteString("## " + renderRich(b.Heading2.RichText) + "\n\n")
		case "heading_3":
			sb.WriteString("### " + renderRich(b.Heading3.RichText) + "\n\n")
		case "paragraph":
			sb.WriteString(renderRich(b.Paragraph.RichText) + "\n\n")
		case "bulleted_list_item":
			sb.WriteString("- " + renderRich(b.Bulleted.RichText) + "\n")
		case "numbered_list_item":
			numberCounter++
			sb.WriteString(itoa(numberCounter) + ". " + renderRich(b.Numbered.RichText) + "\n")
		case "to_do":
			mark := "[ ]"
			if b.ToDo.Checked {
				mark = "[x]"
			}
			sb.WriteString("- " + mark + " " + renderRich(b.ToDo.RichText) + "\n")
		case "quote":
			sb.WriteString("> " + renderRich(b.Quote.RichText) + "\n\n")
		case "code":
			sb.WriteString("```" + b.Code.Language + "\n" + plain(b.Code.RichText) + "\n```\n\n")
		case "divider":
			sb.WriteString("---\n\n")
		}
	}
	return sb.String()
}

func renderRich(spans []notionRichText) string {
	var sb strings.Builder
	for _, s := range spans {
		t := s.PlainText
		if s.Annotations.Code {
			t = "`" + t + "`"
		}
		if s.Annotations.Bold {
			t = "**" + t + "**"
		}
		if s.Annotations.Italic {
			t = "*" + t + "*"
		}
		if s.Annotations.Strikethrough {
			t = "~~" + t + "~~"
		}
		if s.Href != "" {
			t = "[" + t + "](" + s.Href + ")"
		}
		sb.WriteString(t)
	}
	return sb.String()
}

func plain(spans []notionRichText) string {
	var sb strings.Builder
	for _, s := range spans {
		sb.WriteString(s.PlainText)
	}
	return sb.String()
}

func itoa(n int) string {
	return strings.TrimSpace(strings.Replace(" 0123456789", " ", "", 1)[:0]) + fmtInt(n)
}

func fmtInt(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
