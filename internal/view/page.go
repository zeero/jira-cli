package view

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/glamour"

	"github.com/ankitpokhrel/jira-cli/pkg/confluence"
	"github.com/ankitpokhrel/jira-cli/pkg/tui"
)

// Page is a view for a Confluence page.
type Page struct {
	Data    *confluence.Page
	Display DisplayFormat
}

// Render renders the page view.
func (p Page) Render() error {
	if p.Display.Plain || tui.IsDumbTerminal() || tui.IsNotTTY() {
		return p.renderPlain(os.Stdout)
	}
	r, err := MDRenderer()
	if err != nil {
		return err
	}
	out, err := p.RenderedOut(r)
	if err != nil {
		return err
	}
	return tui.PagerOut(out)
}

// RenderedOut translates raw data to the format we want to display in.
func (p Page) RenderedOut(renderer *glamour.TermRenderer) (string, error) {
	var res strings.Builder

	for _, fragment := range p.fragments() {
		if fragment.Parse {
			out, err := renderer.Render(fragment.Body)
			if err != nil {
				return "", err
			}
			res.WriteString(out)
		} else {
			res.WriteString(fragment.Body)
		}
	}

	return res.String(), nil
}

func (p Page) String() string {
	var s strings.Builder

	s.WriteString(p.header())

	content := p.content()
	if content != "" {
		s.WriteString(fmt.Sprintf("\n\n%s\n\n%s", p.separator("Content"), content))
	}

	return s.String()
}

func (p Page) fragments() []fragment {
	scraps := []fragment{
		{Body: p.header(), Parse: true},
	}

	content := p.content()
	if content != "" {
		scraps = append(
			scraps,
			newBlankFragment(1),
			fragment{Body: p.separator("Content")},
			newBlankFragment(2),
			// Storage formatはHTMLなので、Parseせずにそのまま表示
			fragment{Body: content, Parse: false},
		)
	}

	return scraps
}

func (p Page) header() string {
	var statusIcon string
	switch p.Data.Status {
	case "current":
		statusIcon = "✅"
	case "draft":
		statusIcon = "📝"
	default:
		statusIcon = "📄"
	}

	return fmt.Sprintf(
		"# %s\n🔑 ID: %s  %s Status: %s  📦 Space ID: %s",
		p.Data.Title,
		p.Data.ID,
		statusIcon,
		p.Data.Status,
		p.Data.SpaceID,
	)
}

func (p Page) content() string {
	if p.Data.Body == nil || p.Data.Body.Storage == nil {
		return ""
	}
	return p.Data.Body.Storage.Value
}

func (p Page) separator(title string) string {
	return fmt.Sprintf("─── %s ───", title)
}

func (p Page) renderPlain(w io.Writer) error {
	_, err := fmt.Fprint(w, p.String())
	return err
}
