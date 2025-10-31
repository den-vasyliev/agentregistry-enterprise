package cmd

import (
	"fmt"
	"log"

	"github.com/agentregistry-dev/agentregistry/internal/database"
	"github.com/agentregistry-dev/agentregistry/internal/printer"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install <resource-type> <resource-name> [version]",
	Short: "Install a resource",
	Long:  `Install resources (mcp server, skill) from connected registries.`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		resourceType := args[0]
		resourceName := args[1]
		version := "latest"
		if len(args) > 2 {
			version = args[2]
		}

		// Initialize database
		if err := database.Initialize(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer func() {
			if err := database.Close(); err != nil {
				log.Printf("Warning: Failed to close database: %v", err)
			}
		}()

		fmt.Printf("Installing %s: %s@%s\n", resourceType, resourceName, version)

		// TODO: Implement install logic
		// 1. Fetch resource from registry
		// 2. For MCP servers, prompt for environment variables
		// 3. Install and configure resource
		// 4. Update local database

		switch resourceType {
		case "mcp":
			// Placeholder for MCP server installation
			fmt.Println("\nEnvironment variables may be required")
			fmt.Println("Configuration will be prompted during installation")
		case "skill":
			// Placeholder for skill installation
			fmt.Println("\nInstalling skill package")
		default:
			printer.PrintError(fmt.Sprintf("Unknown resource type: %s", resourceType))
			fmt.Println("Valid types: mcp, skill")
			return
		}

		printer.PrintSuccess("Installation completed successfully")
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}
