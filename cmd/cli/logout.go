package main

import (
	"fmt"

	"github.com/link-rift/link-rift/internal/cli"
	"github.com/spf13/cobra"
)

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out and clear stored credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cli.LoadConfig()
			if err != nil {
				return err
			}

			if cfg.AccessToken != "" {
				client := cli.NewClient(cfg.APIBaseURL, cfg.AccessToken)
				// Best-effort server-side logout
				_, _ = client.Post("/api/v1/auth/logout", nil)
			}

			if err := cli.ClearAuth(cfg); err != nil {
				return fmt.Errorf("failed to clear credentials: %w", err)
			}

			fmt.Println("Logged out successfully.")
			return nil
		},
	}
}
