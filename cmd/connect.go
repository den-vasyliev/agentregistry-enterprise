package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/agentregistry-dev/agentregistry/internal/database"
	"github.com/spf13/cobra"
)

var connectCmd = &cobra.Command{
	Use:   "connect <registry-url> <registry-name>",
	Short: "Connect to a registry",
	Long:  `Connects an existing registry to arctl. This will fetch the data from the registry and store it locally.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		registryURL := args[0]
		registryName := args[1]

		// Initialize database
		if err := database.Initialize(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer func() {
			if err := database.Close(); err != nil {
				log.Printf("Warning: Failed to close database: %v", err)
			}
		}()

		fmt.Printf("Connecting to registry: %s (%s)\n", registryName, registryURL)

		// Add the registry with default type
		if err := database.AddRegistry(registryName, registryURL, "registry"); err != nil {
			if strings.Contains(err.Error(), "UNIQUE constraint failed") {
				log.Fatalf("Registry '%s' already exists", registryName)
			}
			log.Fatalf("Failed to add registry: %v", err)
		}

		fmt.Println("✓ Registry connected successfully")
		fmt.Println("\nNext steps:")
		fmt.Println("  Run 'arctl refresh' to fetch registry data")
		fmt.Println("  Run 'arctl list mcp' to see available MCP servers")
	},
}

func init() {
	rootCmd.AddCommand(connectCmd)
}
