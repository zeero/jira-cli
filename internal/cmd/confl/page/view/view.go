package view

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	tuiView "github.com/ankitpokhrel/jira-cli/internal/view"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `View displays contents of a Confluence page in storage format (XHTML).`
	examples = `$ confl page view 12345

# Display output in plain mode
$ confl page view 12345 --plain`

	flagPlain = "plain"
)

// NewCmdView is a view command.
func NewCmdView() *cobra.Command {
	cmd := cobra.Command{
		Use:     "view PAGE-ID",
		Short:   "View displays contents of a Confluence page",
		Long:    helpText,
		Example: examples,
		Annotations: map[string]string{
			"help:args": "PAGE-ID\tConfluence page ID, eg: 12345",
		},
		Args: cobra.ExactArgs(1),
		Run:  view,
	}

	cmd.Flags().Bool(flagPlain, false, "Display output in plain mode")

	return &cmd
}

func view(cmd *cobra.Command, args []string) {
	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	plain, err := cmd.Flags().GetBool(flagPlain)
	cmdutil.ExitIfError(err)

	pageID := args[0]

	page, err := func() (*tuiView.Page, error) {
		s := cmdutil.Info("Fetching page...")
		defer s.Stop()

		client := api.ConfluenceClient(jira.Config{Debug: debug})
		p, err := client.GetPage(pageID)
		if err != nil {
			return nil, err
		}

		return &tuiView.Page{
			Data:    p,
			Display: tuiView.DisplayFormat{Plain: plain},
		}, nil
	}()
	cmdutil.ExitIfError(err)

	cmdutil.ExitIfError(page.Render())
}
