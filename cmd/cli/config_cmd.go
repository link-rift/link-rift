package main

import (
	"fmt"

	"github.com/link-rift/link-rift/internal/cli"
	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}

	cmd.AddCommand(
		newConfigShowCmd(),
		newConfigSetURLCmd(),
		newConfigSetWorkspaceCmd(),
	)

	return cmd
}

func newConfigShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cli.LoadConfig()
			if err != nil {
				return err
			}

			fmt.Printf("API Base URL:  %s\n", cfg.APIBaseURL)
			fmt.Printf("Workspace ID:  %s\n", valueOrNone(cfg.WorkspaceID))
			fmt.Printf("Authenticated: %v\n", cfg.AccessToken != "")
			fmt.Printf("Config file:   %s\n", cli.ConfigPath())
			return nil
		},
	}
}

func newConfigSetURLCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-url <url>",
		Short: "Set the API base URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cli.LoadConfig()
			if err != nil {
				return err
			}
			cfg.APIBaseURL = args[0]
			if err := cli.SaveConfig(cfg); err != nil {
				return err
			}
			fmt.Printf("API URL set to %s\n", args[0])
			return nil
		},
	}
}

func newConfigSetWorkspaceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set-workspace <workspace-id>",
		Short: "Set the active workspace",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cli.LoadConfig()
			if err != nil {
				return err
			}
			cfg.WorkspaceID = args[0]
			if err := cli.SaveConfig(cfg); err != nil {
				return err
			}
			fmt.Printf("Workspace set to %s\n", args[0])
			return nil
		},
	}
}

func valueOrNone(s string) string {
	if s == "" {
		return "(none)"
	}
	return s
}
