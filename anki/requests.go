package anki

type AddNoteRequest struct {
	Action  string `json:"action,omitempty"`
	Version int    `json:"version,omitempty"`
	Params  Params `json:"params,omitempty"`
}

type Params struct {
	Note Note `json:"note,omitempty"`
}

type Note struct {
	DeckName  string            `json:"deckName,omitempty"`
	ModelName string            `json:"modelName,omitempty"`
	Fields    map[string]string `json:"fields,omitempty"`
	Options   map[string]any    `json:"options,omitempty"`
	Tags      []string          `json:"tags,omitempty"`
}
