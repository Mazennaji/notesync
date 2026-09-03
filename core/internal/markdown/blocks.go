package markdown

type RichText struct {
	Content string
	Bold    bool
	Italic  bool
	Code    bool
	Strike  bool
	Link    string
}

type BlockType string

const (
	Paragraph  BlockType = "paragraph"
	Heading1   BlockType = "heading_1"
	Heading2   BlockType = "heading_2"
	Heading3   BlockType = "heading_3"
	BulletItem BlockType = "bulleted_list_item"
	NumberItem BlockType = "numbered_list_item"
	ToDoItem   BlockType = "to_do"
	Quote      BlockType = "quote"
	Code       BlockType = "code"
	Divider    BlockType = "divider"
)

type Block struct {
	Type     BlockType
	Text     []RichText
	Checked  bool
	Language string
}
