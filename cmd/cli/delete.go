package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/link-rift/link-rift/internal/cli"
	"github.com/spf13/cobra"
)

func newDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete <link-id>",
		Aliases: []string{"rm"},
		Short:   "Delete a shortened link",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Delete link %s? [y/N] ", args[0])
				reader := bufio.NewReader(os.Stdin)
				input, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(input)) != "y" {
					fmt.Println("Cancelled.")
					return nil
				}
			}

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

			path := fmt.Sprintf("/api/v1/workspaces/%s/links/%s", cfg.WorkspaceID, args[0])
			_, err = client.Delete(path)
			if err != nil {
				return err
			}

			fmt.Println("Link deleted.")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation prompt")

	return cmd
}
