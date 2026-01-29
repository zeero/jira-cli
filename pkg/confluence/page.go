package confluence

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

// GetPage fetches a Confluence page by ID with storage format body.
func (c *Client) GetPage(pageID string) (*Page, error) {
	path := fmt.Sprintf("/pages/%s?body-format=storage", pageID)

	res, err := c.Get(context.Background(), path, jira.Header{
		"Accept": "application/json",
	})
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("confluence: unexpected response '%s'", res.Status)
	}

	var page Page
	if err := json.NewDecoder(res.Body).Decode(&page); err != nil {
		return nil, err
	}

	return &page, nil
}
