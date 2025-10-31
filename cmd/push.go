package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/agentregistry-dev/agentregistry/internal/database"
	"github.com/agentregistry-dev/agentregistry/internal/detector"
	"github.com/agentregistry-dev/agentregistry/internal/docker"
	"github.com/spf13/cobra"
)

var (
	pushName     string
	pushRegistry string
	pushTag      string
	pushPlatform string
)

var pushCmd = &cobra.Command{
	Use:   "push [directory]",
	Short: "Package and push an MCP server to a registry",
	Long: `Package an MCP server as a Docker image and push it to a container registry.

This command will:
1. Detect the MCP server type (npm or uv/Python based)
2. Generate a Dockerfile if one doesn't exist
3. Build a Docker image
4. Push the image to the configured registry
5. Register the server metadata in the local database

Examples:
  # Push current directory with auto-detected name
  arctl push .

  # Push with custom name
  arctl push -n my-weather-server .

  # Push to a specific registry
  arctl push -r ghcr.io/username -n my-server .

  # Push with custom tag
  arctl push -n my-server -t v1.0.0 .`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		directory := args[0]

		// Check Docker is installed
		if err := docker.CheckDockerInstalled(); err != nil {
			log.Fatalf("Error: %v", err)
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

		fmt.Printf("Detecting MCP server in: %s\n", directory)

		// Detect MCP server type
		serverInfo, err := detector.DetectMCPServer(directory)
		if err != nil {
			log.Fatalf("Failed to detect MCP server: %v", err)
		}

		fmt.Printf("✓ Detected %s-based MCP server\n", serverInfo.Type)
		if serverInfo.Name != "" {
			fmt.Printf("  Name: %s\n", serverInfo.Name)
		}
		if serverInfo.Version != "" {
			fmt.Printf("  Version: %s\n", serverInfo.Version)
		}
		if serverInfo.Description != "" {
			fmt.Printf("  Description: %s\n", serverInfo.Description)
		}
		if serverInfo.Framework != "" {
			fmt.Printf("  Framework: %s\n", serverInfo.Framework)
		}
		if serverInfo.PackageManager != "" {
			fmt.Printf("  Package Manager: %s\n", serverInfo.PackageManager)
		}
		if serverInfo.EntryPoint != "" {
			fmt.Printf("  Entry Point: %s\n", serverInfo.EntryPoint)
		}
		if serverInfo.HasDockerfile {
			fmt.Printf("  Dockerfile: ✓ Found\n")
		}

		// Determine image name
		imageName := pushName
		if imageName == "" {
			if serverInfo.Name != "" {
				imageName = serverInfo.Name
			} else {
				log.Fatal("Error: Could not determine server name. Use -n flag to specify a name")
			}
		}

		// Determine tag
		imageTag := pushTag
		if imageTag == "" {
			if serverInfo.Version != "" {
				imageTag = serverInfo.Version
			} else {
				imageTag = "latest"
			}
		}

		// Validate registry is set
		if pushRegistry == "" {
			pushRegistry = GetDefaultRegistry()
			if pushRegistry == "" {
				log.Fatal("Error: Registry not configured. Use -r flag or set default: arctl config registry <registry>")
			}
		}

		// Check Docker login
		if err := docker.CheckDockerLogin(pushRegistry); err != nil {
			log.Printf("Warning: %v", err)
		}

		// Build Docker image
		buildConfig := &docker.BuildConfig{
			ServerInfo: serverInfo,
			ImageName:  imageName,
			ImageTag:   imageTag,
			Registry:   pushRegistry,
			Platform:   pushPlatform,
		}

		builder := docker.NewBuilder(buildConfig)

		fmt.Println("\n📦 Building Docker image...")
		if err := builder.Build(); err != nil {
			log.Fatalf("Failed to build Docker image: %v", err)
		}

		fmt.Println("\n🚀 Pushing to registry...")
		if err := builder.Push(); err != nil {
			log.Fatalf("Failed to push Docker image: %v", err)
		}

		// Create server metadata for database
		fmt.Println("\n💾 Registering server metadata...")
		if err := registerServerMetadata(serverInfo, builder.GetImageReference(), imageName, imageTag); err != nil {
			log.Fatalf("Failed to register server metadata: %v", err)
		}

		fmt.Println("\n✅ MCP server pushed successfully!")
		fmt.Printf("\nImage: %s\n", builder.GetImageReference())
		fmt.Printf("\nTo install this server:\n")
		fmt.Printf("  arctl pull %s\n", imageName)
		fmt.Printf("  arctl install mcp %s\n", imageName)
	},
}

func registerServerMetadata(serverInfo *detector.MCPServerInfo, imageRef, name, tag string) error {
	// Get or create a "local" registry for pushed images
	registries, err := database.GetRegistries()
	if err != nil {
		return fmt.Errorf("failed to get registries: %w", err)
	}

	var localRegistryID int
	for _, reg := range registries {
		if reg.Name == "local" {
			localRegistryID = reg.ID
			break
		}
	}

	if localRegistryID == 0 {
		// Create local registry
		if err := database.AddRegistry("local", "local://", "private"); err != nil {
			if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
				return fmt.Errorf("failed to create local registry: %w", err)
			}
			// If we hit unique constraint, try to get it again
			registries, err := database.GetRegistries()
			if err != nil {
				return fmt.Errorf("failed to get registries after creation: %w", err)
			}
			for _, reg := range registries {
				if reg.Name == "local" {
					localRegistryID = reg.ID
					break
				}
			}
		} else {
			// Get the ID of newly created registry
			registries, err := database.GetRegistries()
			if err != nil {
				return fmt.Errorf("failed to get registries: %w", err)
			}
			for _, reg := range registries {
				if reg.Name == "local" {
					localRegistryID = reg.ID
					break
				}
			}
		}
	}

	// Create server metadata
	description := serverInfo.Description
	if description == "" {
		description = fmt.Sprintf("MCP server pushed from local machine (%s-based)", serverInfo.Type)
	}

	// Add framework info if available (for kmcp servers)
	if serverInfo.Framework != "" {
		description = fmt.Sprintf("%s (framework: %s)", description, serverInfo.Framework)
	}

	serverData := map[string]interface{}{
		"name":        name,
		"title":       name,
		"description": description,
		"version":     tag,
		"packages": []map[string]interface{}{
			{
				"identifier":   imageRef,
				"version":      tag,
				"registryType": "docker",
				"transport": map[string]string{
					"type": "stdio",
				},
			},
		},
	}

	// Add framework metadata if available
	if serverInfo.Framework != "" {
		serverData["framework"] = serverInfo.Framework
	}

	serverDataJSON, err := json.Marshal(serverData)
	if err != nil {
		return fmt.Errorf("failed to marshal server data: %w", err)
	}

	// Add or update server in database
	title := name

	if err := database.AddOrUpdateServer(
		localRegistryID,
		name,
		title,
		description,
		tag,
		"",
		string(serverDataJSON),
	); err != nil {
		return fmt.Errorf("failed to add server to database: %w", err)
	}

	fmt.Printf("✓ Registered server in local registry\n")
	return nil
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.Flags().StringVarP(&pushName, "name", "n", "", "Name for the MCP server (auto-detected if not specified)")
	pushCmd.Flags().StringVarP(&pushRegistry, "registry", "r", "", "Container registry (e.g., docker.io/username, ghcr.io/username)")
	pushCmd.Flags().StringVarP(&pushTag, "tag", "t", "", "Image tag (uses version from package file if available, otherwise 'latest')")
	pushCmd.Flags().StringVarP(&pushPlatform, "platform", "p", "", "Target platform(s) (e.g., linux/amd64,linux/arm64)")
}
