package root

import (
	"fmt"
	"os"
	"slices"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/internal/cmd/confl/page"
	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	jiraConfig "github.com/ankitpokhrel/jira-cli/internal/config"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
	"github.com/ankitpokhrel/jira-cli/pkg/netrc"

	"github.com/zalando/go-keyring"
)

const (
	jiraCLIHelpLink  = "https://github.com/ankitpokhrel/jira-cli#getting-started"
	jiraAPITokenLink = "https://id.atlassian.com/manage-profile/security/api-tokens"
)

var (
	config string
	debug  bool
)

func init() {
	cobra.OnInitialize(func() {
		if config != "" {
			viper.SetConfigFile(config)
		} else if configFile := os.Getenv("JIRA_CONFIG_FILE"); configFile != "" {
			viper.SetConfigFile(configFile)
		} else {
			home, err := cmdutil.GetConfigHome()
			if err != nil {
				cmdutil.Failed("Error: %s", err)
				return
			}

			viper.AddConfigPath(fmt.Sprintf("%s/%s", home, jiraConfig.Dir))
			viper.SetConfigName(jiraConfig.FileName)
			viper.SetConfigType(jiraConfig.FileType)
		}

		viper.AutomaticEnv()
		viper.SetEnvPrefix("jira")

		if err := viper.ReadInConfig(); err == nil && debug {
			fmt.Printf("Using config file: %s\n", viper.ConfigFileUsed())
		}
	})
}

// NewCmdRoot is the root command for confl.
func NewCmdRoot() *cobra.Command {
	cmd := cobra.Command{
		Use:   "confl <command> <subcommand>",
		Short: "Confluence CLI",
		Long:  "Interactive Confluence command line.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		PersistentPreRun: func(cmd *cobra.Command, _ []string) {
			subCmd := cmd.Name()
			if !cmdRequireToken(subCmd) {
				return
			}

			if viper.GetString("auth_type") != string(jira.AuthTypeMTLS) {
				checkForToken(viper.GetString("server"), viper.GetString("login"))
			}

			configFile := viper.ConfigFileUsed()
			if !jiraConfig.Exists(configFile) {
				cmdutil.Failed("Missing configuration file.\nRun 'jira init' to configure the tool.")
			}
		},
	}

	configHome, err := cmdutil.GetConfigHome()
	if err != nil {
		cmdutil.Failed("Error: %s", err)
	}

	cmd.PersistentFlags().StringVarP(
		&config, "config", "c", "",
		fmt.Sprintf("Config file (default is %s/%s/%s.yml)", configHome, jiraConfig.Dir, jiraConfig.FileName),
	)
	cmd.PersistentFlags().BoolVar(&debug, "debug", false, "Turn on debug output")

	_ = viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config"))
	_ = viper.BindPFlag("debug", cmd.PersistentFlags().Lookup("debug"))

	cmd.AddCommand(page.NewCmdPage())

	return &cmd
}

func cmdRequireToken(cmd string) bool {
	allowList := []string{
		"help",
		"confl",
		"__complete", "__completeNoDesc",
	}
	return !slices.Contains(allowList, cmd)
}

func checkForToken(server string, login string) {
	if os.Getenv("JIRA_API_TOKEN") != "" {
		return
	}

	netrcConfig, _ := netrc.Read(server, login)
	if netrcConfig != nil {
		return
	}

	secret, _ := keyring.Get("jira-cli", login)
	if secret != "" {
		return
	}

	msg := fmt.Sprintf(`The tool needs a Jira API token to function.

For cloud server: you can generate the token using this link: %s
For local server: you can use the password you use to log in to Jira for basic auth or get a token from your Jira profile for PAT.

After generating the token, you can either:
  - Export API token to your shell as a JIRA_API_TOKEN env variable
  - Or, you can use a .netrc file to define required machine details

Once you are done with the above steps, run 'jira init' to generate the config if you haven't already.

For more details, see: %s
`, jiraAPITokenLink, jiraCLIHelpLink)

	cmdutil.Warn(msg)
	os.Exit(1)
}
