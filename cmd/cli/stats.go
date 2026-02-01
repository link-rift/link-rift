package main

import (
	"encoding/json"
	"fmt"

	"github.com/link-rift/link-rift/internal/cli"
	"github.com/spf13/cobra"
)

type quickStats struct {
	TotalClicks  int64  `json:"total_clicks"`
	UniqueClicks int64  `json:"unique_clicks"`
	Clicks24h    int64  `json:"clicks_24h"`
	Clicks7d     int64  `json:"clicks_7d"`
	CreatedAt    string `json:"created_at"`
}

func newStatsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stats <link-id>",
		Short: "Show quick stats for a link",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := cli.NewClientFromConfig()
			if err != nil {
				return err
			}

			cfg, err := cli.LoadConfig()
			if err != nil {
				return err
			}

			if cfg.WorkspaceID == "" {
				return fmt.Errorf("no workspace selected — run `linkrift config set-workspace <id>` first")
			}

			path := fmt.Sprintf("/api/v1/workspaces/%s/links/%s/stats", cfg.WorkspaceID, args[0])
			resp, err := client.Get(path)
			if err != nil {
				return err
			}

			var stats quickStats
			if err := json.Unmarshal(resp.Data, &stats); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			fmt.Printf("Link Stats (created %s)\n", stats.CreatedAt)
			fmt.Printf("  Total clicks:  %d\n", stats.TotalClicks)
			fmt.Printf("  Unique clicks: %d\n", stats.UniqueClicks)
			fmt.Printf("  Last 24h:      %d\n", stats.Clicks24h)
			fmt.Printf("  Last 7 days:   %d\n", stats.Clicks7d)

			return nil
		},
	}
}
