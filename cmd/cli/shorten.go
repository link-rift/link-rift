package main

import (
	"encoding/json"
	"fmt"

	"github.com/link-rift/link-rift/internal/cli"
	"github.com/spf13/cobra"
)

type linkResponse struct {
	ID        string `json:"id"`
	ShortCode string `json:"short_code"`
	ShortURL  string `json:"short_url"`
	URL       string `json:"url"`
	Title     string `json:"title,omitempty"`
}

func newShortenCmd() *cobra.Command {
	var title, customCode string

	cmd := &cobra.Command{
		Use:   "shorten <url>",
		Short: "Shorten a URL",
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

			body := map[string]any{
				"url": args[0],
			}
			if title != "" {
				body["title"] = title
			}
			if customCode != "" {
				body["short_code"] = customCode
			}

			path := fmt.Sprintf("/api/v1/workspaces/%s/links", cfg.WorkspaceID)
			resp, err := client.Post(path, body)
			if err != nil {
				return err
			}

			var link linkResponse
			if err := json.Unmarshal(resp.Data, &link); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			fmt.Println(link.ShortURL)
			return nil
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "Link title")
	cmd.Flags().StringVarP(&customCode, "code", "c", "", "Custom short code")

	return cmd
}
