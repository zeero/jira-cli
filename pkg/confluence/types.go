package confluence

// Page represents a Confluence page from the v2 API.
type Page struct {
	ID      string    `json:"id"`
	Status  string    `json:"status"`
	Title   string    `json:"title"`
	SpaceID string    `json:"spaceId"`
	Body    *PageBody `json:"body,omitempty"`
}

// PageBody holds the body representations of a page.
type PageBody struct {
	Storage *StorageBody `json:"storage,omitempty"`
}

// StorageBody holds the storage format body of a page.
type StorageBody struct {
	Representation string `json:"representation"`
	Value          string `json:"value"`
}
