package page

import (
	"github.com/spf13/cobra"

	"github.com/ankitpokhrel/jira-cli/internal/cmd/confl/page/view"
)

const helpText = `Page manages Confluence pages. See available commands below.`

// NewCmdPage is a page command.
func NewCmdPage() *cobra.Command {
	cmd := cobra.Command{
		Use:   "page",
		Short: "Manage Confluence pages",
		Long:  helpText,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(view.NewCmdView())

	return &cmd
}
