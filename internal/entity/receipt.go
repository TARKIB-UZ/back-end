package entity

type Section struct {
	Type    string `json:"type"`
	Content string `json:"content,omitempty"`
	URL     string `json:"url,omitempty"`
}

type Recipe struct {
	Sections []Section `json:"sections"`
}
