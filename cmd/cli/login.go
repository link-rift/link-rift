package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/link-rift/link-rift/internal/cli"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type authResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         userResponse `json:"user"`
}

type userResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func newLoginCmd() *cobra.Command {
	var email, password string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the Linkrift API",
		Long: `Log in to your Linkrift account. You can provide credentials via flags
or enter them interactively when prompted.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			reader := bufio.NewReader(os.Stdin)

			if email == "" {
				fmt.Print("Email: ")
				input, _ := reader.ReadString('\n')
				email = strings.TrimSpace(input)
			}

			if password == "" {
				fmt.Print("Password: ")
				pw, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					return fmt.Errorf("failed to read password: %w", err)
				}
				fmt.Println()
				password = string(pw)
			}

			cfg, err := cli.LoadConfig()
			if err != nil {
				return err
			}

			client := cli.NewClient(cfg.APIBaseURL, "")
			resp, err := client.Post("/api/v1/auth/login", map[string]string{
				"email":    email,
				"password": password,
			})
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			var auth authResponse
			if err := json.Unmarshal(resp.Data, &auth); err != nil {
				return fmt.Errorf("failed to parse login response: %w", err)
			}

			cfg.AccessToken = auth.AccessToken
			cfg.RefreshToken = auth.RefreshToken
			if err := cli.SaveConfig(cfg); err != nil {
				return fmt.Errorf("failed to save credentials: %w", err)
			}

			fmt.Printf("Logged in as %s (%s)\n", auth.User.Name, auth.User.Email)
			return nil
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Account email")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Account password")

	return cmd
}
