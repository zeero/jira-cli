package view

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

const (
	helpText = `View displays contents of a Confluence page in storage format (XHTML).`
	examples = `$ confl page view 12345`
)

// NewCmdView is a view command.
func NewCmdView() *cobra.Command {
	return &cobra.Command{
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
}

func view(cmd *cobra.Command, args []string) {
	debug, err := cmd.Flags().GetBool("debug")
	cmdutil.ExitIfError(err)

	pageID := args[0]

	page, err := func() (string, error) {
		s := cmdutil.Info("Fetching page...")
		defer s.Stop()

		client := api.ConfluenceClient(jira.Config{Debug: debug})
		p, err := client.GetPage(pageID)
		if err != nil {
			return "", err
		}
		if p.Body == nil || p.Body.Storage == nil {
			return "", fmt.Errorf("confluence: page body is empty")
		}
		return p.Body.Storage.Value, nil
	}()
	cmdutil.ExitIfError(err)

	fmt.Println(page)
}
