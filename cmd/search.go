package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/agentregistry-dev/agentregistry/internal/database"
	"github.com/agentregistry-dev/agentregistry/internal/models"
	"github.com/agentregistry-dev/agentregistry/internal/printer"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <resource-type> <search-term>",
	Short: "Search for resources",
	Long:  `Search for resources from the connected registries.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resourceType := args[0]
		searchTerm := strings.ToLower(args[1])

		// Initialize database
		if err := database.Initialize(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer func() {
			if err := database.Close(); err != nil {
				log.Printf("Warning: Failed to close database: %v", err)
			}
		}()

		switch resourceType {
		case "mcp":
			servers, err := database.GetServers()
			if err != nil {
				log.Fatalf("Failed to get servers: %v", err)
			}

			// Filter servers by search term
			var matches []models.ServerDetail
			for _, s := range servers {
				if strings.Contains(strings.ToLower(s.Name), searchTerm) ||
					strings.Contains(strings.ToLower(s.Title), searchTerm) {
					matches = append(matches, s)
				}
			}

			if len(matches) == 0 {
				fmt.Printf("No MCP servers found matching '%s'\n", searchTerm)
				return
			}

			t := printer.NewTablePrinter(os.Stdout)
			t.SetHeaders("Name", "Title", "Version", "Status")

			for _, s := range matches {
				title := printer.EmptyValueOrDefault(s.Title, "<none>")
				status := printer.FormatStatus(s.Installed)

				t.AddRow(
					printer.TruncateString(s.Name, 40),
					printer.TruncateString(title, 30),
					s.Version,
					status,
				)
			}
			t.Render()

		case "skill":
			skills, err := database.GetSkills()
			if err != nil {
				log.Fatalf("Failed to get skills: %v", err)
			}

			// Filter skills by search term
			var matches []models.Skill
			for _, s := range skills {
				if strings.Contains(strings.ToLower(s.Name), searchTerm) ||
					strings.Contains(strings.ToLower(s.Description), searchTerm) {
					matches = append(matches, s)
				}
			}

			if len(matches) == 0 {
				fmt.Printf("No skills found matching '%s'\n", searchTerm)
				return
			}

			t := printer.NewTablePrinter(os.Stdout)
			t.SetHeaders("Name", "Description", "Version", "Status")

			for _, s := range matches {
				status := printer.FormatStatus(s.Installed)

				t.AddRow(
					printer.TruncateString(s.Name, 40),
					printer.TruncateString(s.Description, 50),
					s.Version,
					status,
				)
			}
			t.Render()

		case "registry":
			registries, err := database.GetRegistries()
			if err != nil {
				log.Fatalf("Failed to get registries: %v", err)
			}

			// Filter registries by search term
			var matches []models.Registry
			for _, r := range registries {
				if strings.Contains(strings.ToLower(r.Name), searchTerm) ||
					strings.Contains(strings.ToLower(r.URL), searchTerm) {
					matches = append(matches, r)
				}
			}

			if len(matches) == 0 {
				fmt.Printf("No registries found matching '%s'\n", searchTerm)
				return
			}

			t := printer.NewTablePrinter(os.Stdout)
			t.SetHeaders("Name", "URL", "Type", "Age")

			for _, r := range matches {
				t.AddRow(
					r.Name,
					r.URL,
					r.Type,
					printer.FormatAge(r.CreatedAt),
				)
			}
			t.Render()

		default:
			fmt.Printf("Unknown resource type: %s\n", resourceType)
			fmt.Println("Valid types: mcp, skill, registry")
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
