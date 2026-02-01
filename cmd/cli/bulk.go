package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/link-rift/link-rift/internal/cli"
	"github.com/spf13/cobra"
)

func newBulkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bulk",
		Short: "Bulk operations",
	}

	cmd.AddCommand(newBulkImportCmd())
	return cmd
}

type bulkLink struct {
	URL       string `json:"url"`
	ShortCode string `json:"short_code,omitempty"`
	Title     string `json:"title,omitempty"`
}

func newBulkImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import <file>",
		Short: "Bulk import links from a CSV or JSON file",
		Long: `Import links in bulk from a file. Supported formats:

CSV: Header row with columns: url, short_code (optional), title (optional)
JSON: Array of objects with fields: url, short_code (optional), title (optional)`,
		Args: cobra.ExactArgs(1),
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

			filePath := args[0]
			links, err := parseBulkFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to parse file: %w", err)
			}

			if len(links) == 0 {
				return fmt.Errorf("no links found in file")
			}

			if len(links) > 100 {
				return fmt.Errorf("maximum 100 links per import (got %d)", len(links))
			}

			fmt.Printf("Importing %d links...\n", len(links))

			path := fmt.Sprintf("/api/v1/workspaces/%s/links/bulk", cfg.WorkspaceID)
			resp, err := client.Post(path, map[string]any{
				"links": links,
			})
			if err != nil {
				return err
			}

			var created []linkResponse
			if err := json.Unmarshal(resp.Data, &created); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			fmt.Printf("Successfully imported %d links.\n", len(created))
			for _, l := range created {
				fmt.Printf("  %s → %s\n", l.ShortURL, l.URL)
			}

			return nil
		},
	}
}

func parseBulkFile(filePath string) ([]bulkLink, error) {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".json":
		return parseBulkJSON(filePath)
	case ".csv":
		return parseBulkCSV(filePath)
	default:
		return nil, fmt.Errorf("unsupported file format %q (use .csv or .json)", ext)
	}
}

func parseBulkJSON(filePath string) ([]bulkLink, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var links []bulkLink
	if err := json.Unmarshal(data, &links); err != nil {
		return nil, err
	}
	return links, nil
}

func parseBulkCSV(filePath string) ([]bulkLink, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file must have a header row and at least one data row")
	}

	header := records[0]
	colIdx := map[string]int{}
	for i, col := range header {
		colIdx[strings.ToLower(strings.TrimSpace(col))] = i
	}

	urlIdx, ok := colIdx["url"]
	if !ok {
		return nil, fmt.Errorf("CSV must have a 'url' column")
	}

	var links []bulkLink
	for _, row := range records[1:] {
		if urlIdx >= len(row) || strings.TrimSpace(row[urlIdx]) == "" {
			continue
		}

		link := bulkLink{URL: strings.TrimSpace(row[urlIdx])}

		if idx, ok := colIdx["short_code"]; ok && idx < len(row) {
			link.ShortCode = strings.TrimSpace(row[idx])
		}
		if idx, ok := colIdx["title"]; ok && idx < len(row) {
			link.Title = strings.TrimSpace(row[idx])
		}

		links = append(links, link)
	}

	return links, nil
}
