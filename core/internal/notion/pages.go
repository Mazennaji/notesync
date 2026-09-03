package notion

import (
	"encoding/json"
	"fmt"
)

type Page struct {
	ID    string
	Title string
}

type searchRequest struct {
	Filter      searchFilter `json:"filter"`
	StartCursor string       `json:"start_cursor,omitempty"`
	PageSize    int          `json:"page_size"`
}

type searchFilter struct {
	Property string `json:"property"`
	Value    string `json:"value"`
}

type searchResponse struct {
	Results    []json.RawMessage `json:"results"`
	NextCursor string            `json:"next_cursor"`
	HasMore    bool              `json:"has_more"`
}

type createPageRequest struct {
	Parent     pageParent               `json:"parent"`
	Properties map[string]titleProperty `json:"properties"`
}

type pageParent struct {
	PageID string `json:"page_id"`
}

type titleProperty struct {
	Title []richText `json:"title"`
}

type richText struct {
	Text textContent `json:"text"`
}

type textContent struct {
	Content string `json:"content"`
}

func (c *Client) CreatePage(parentID, title string) (string, error) {
	body := createPageRequest{
		Parent: pageParent{PageID: parentID},
		Properties: map[string]titleProperty{
			"title": {Title: []richText{{Text: textContent{Content: title}}}},
		},
	}

	data, err := c.do("POST", "/pages", body)
	if err != nil {
		return "", err
	}

	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", err
	}
	return resp.ID, nil
}

func (c *Client) SearchPages() ([]Page, error) {
	var pages []Page
	cursor := ""

	for {
		reqBody := searchRequest{
			Filter:      searchFilter{Property: "object", Value: "page"},
			StartCursor: cursor,
			PageSize:    100,
		}

		data, err := c.do("POST", "/search", reqBody)
		if err != nil {
			return nil, err
		}

		var sr searchResponse
		if err := json.Unmarshal(data, &sr); err != nil {
			return nil, err
		}

		for _, raw := range sr.Results {
			if p, ok := parsePage(raw); ok {
				pages = append(pages, p)
			}
		}

		if !sr.HasMore || sr.NextCursor == "" {
			break
		}
		cursor = sr.NextCursor
	}

	return pages, nil
}

func parsePage(raw json.RawMessage) (Page, bool) {
	var obj struct {
		ID         string                     `json:"id"`
		Properties map[string]json.RawMessage `json:"properties"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil || obj.ID == "" {
		return Page{}, false
	}

	return Page{ID: obj.ID, Title: extractTitle(obj.Properties)}, true
}

func extractTitle(props map[string]json.RawMessage) string {
	for _, raw := range props {
		var prop struct {
			Type  string `json:"type"`
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		}
		if err := json.Unmarshal(raw, &prop); err != nil {
			continue
		}
		if prop.Type == "title" {
			title := ""
			for _, t := range prop.Title {
				title += t.PlainText
			}
			if title != "" {
				return title
			}
		}
	}
	return "Untitled"
}

func (c *Client) CheckPage(pageID string) error {
	_, err := c.do("GET", "/pages/"+pageID, nil)
	if err != nil {
		return fmt.Errorf("parent page not accessible (shared with integration?): %w", err)
	}
	return nil
}
