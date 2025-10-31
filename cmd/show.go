package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/agentregistry-dev/agentregistry/internal/database"
	"github.com/agentregistry-dev/agentregistry/internal/printer"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <resource-type> <resource-name>",
	Short: "Show details of a resource",
	Long:  `Shows detailed information about a resource (mcp, skill, registry).`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resourceType := args[0]
		resourceName := args[1]

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
			server, err := database.GetServerByName(resourceName)
			if err != nil {
				log.Fatalf("Failed to get server: %v", err)
			}
			if server == nil {
				fmt.Printf("Server '%s' not found\n", resourceName)
				return
			}

			// Display server details in table format
			t := printer.NewTablePrinter(os.Stdout)
			t.SetHeaders("Property", "Value")
			t.AddRow("Name", server.Name)
			t.AddRow("Title", printer.EmptyValueOrDefault(server.Title, "<none>"))
			t.AddRow("Version", server.Version)
			t.AddRow("Status", printer.FormatStatus(server.Installed))
			t.AddRow("Registry", server.RegistryName)
			t.Render()

		case "skill":
			skill, err := database.GetSkillByName(resourceName)
			if err != nil {
				log.Fatalf("Failed to get skill: %v", err)
			}
			if skill == nil {
				fmt.Printf("Skill '%s' not found\n", resourceName)
				return
			}

			// Display skill details in table format
			t := printer.NewTablePrinter(os.Stdout)
			t.SetHeaders("Property", "Value")
			t.AddRow("Name", skill.Name)
			t.AddRow("Description", skill.Description)
			t.AddRow("Version", skill.Version)
			t.AddRow("Status", printer.FormatStatus(skill.Installed))
			t.AddRow("Registry", skill.RegistryName)
			t.Render()

		case "registry":
			registry, err := database.GetRegistryByName(resourceName)
			if err != nil {
				log.Fatalf("Failed to get registry: %v", err)
			}
			if registry == nil {
				fmt.Printf("Registry '%s' not found\n", resourceName)
				return
			}

			// Display registry details in table format
			t := printer.NewTablePrinter(os.Stdout)
			t.SetHeaders("Property", "Value")
			t.AddRow("Name", registry.Name)
			t.AddRow("URL", registry.URL)
			t.AddRow("Type", registry.Type)
			t.AddRow("Added", printer.FormatTimestampShort(registry.CreatedAt))
			t.AddRow("Age", printer.FormatAge(registry.CreatedAt))
			t.Render()

		default:
			fmt.Printf("Unknown resource type: %s\n", resourceType)
			fmt.Println("Valid types: mcp, skill, registry")
		}
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
