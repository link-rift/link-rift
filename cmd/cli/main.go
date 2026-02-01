package main

import (
	"fmt"
	"os"

	"github.com/link-rift/link-rift/internal/cli"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "linkrift",
		Short: "Linkrift CLI — manage links from the command line",
		Long: `Linkrift CLI lets you shorten URLs, manage links, and view analytics
from your terminal. Use 'linkrift login' to authenticate, then start
shortening URLs with 'linkrift shorten <url>'.`,
		Version:       fmt.Sprintf("%s (commit: %s)", version, commit),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.PersistentFlags().String("api-url", "", "Override API base URL")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		apiURL, _ := cmd.Flags().GetString("api-url")
		if apiURL != "" {
			cfg, err := cli.LoadConfig()
			if err != nil {
				return err
			}
			cfg.APIBaseURL = apiURL
			if err := cli.SaveConfig(cfg); err != nil {
				return err
			}
		}
		return nil
	}

	rootCmd.AddCommand(
		newLoginCmd(),
		newLogoutCmd(),
		newShortenCmd(),
		newListCmd(),
		newStatsCmd(),
		newDeleteCmd(),
		newBulkCmd(),
		newConfigCmd(),
	)

	rootCmd.InitDefaultCompletionCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
