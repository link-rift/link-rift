package main

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"os"

	"github.com/link-rift/link-rift/internal/cli"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var limit, offset int
	var search string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List shortened links",
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

			path := fmt.Sprintf("/api/v1/workspaces/%s/links?limit=%d&offset=%d", cfg.WorkspaceID, limit, offset)
			if search != "" {
				path += "&search=" + search
			}

			resp, err := client.Get(path)
			if err != nil {
				return err
			}

			var links []linkResponse
			if err := json.Unmarshal(resp.Data, &links); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			if len(links) == 0 {
				fmt.Println("No links found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SHORT CODE\tSHORT URL\tDESTINATION\tTITLE")
			fmt.Fprintln(w, "----------\t---------\t-----------\t-----")
			for _, l := range links {
				dest := l.URL
				if len(dest) > 60 {
					dest = dest[:57] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", l.ShortCode, l.ShortURL, dest, l.Title)
			}
			w.Flush()

			if resp.Meta != nil {
				fmt.Printf("\nShowing %d of %d links\n", len(links), resp.Meta.Total)
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "l", 20, "Number of links to show (max 100)")
	cmd.Flags().IntVarP(&offset, "offset", "o", 0, "Offset for pagination")
	cmd.Flags().StringVarP(&search, "search", "s", "", "Search filter")

	return cmd
}
