package cmd

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/agentregistry-dev/agentregistry/internal/database"
	"github.com/agentregistry-dev/agentregistry/internal/docker"
	"github.com/spf13/cobra"
)

var (
	pullPlatform string
)

var pullCmd = &cobra.Command{
	Use:   "pull <server-name>",
	Short: "Pull an MCP server Docker image to local machine",
	Long: `Pull an MCP server Docker image from a container registry to your local machine.

This command will:
1. Look up the server in the local database
2. Get the Docker image reference
3. Pull the image using Docker

The server must be registered in your local database (either from a connected
registry or previously pushed using 'arctl push').

Examples:
  # Pull a server by name
  arctl pull my-weather-server

  # Pull for specific platform
  arctl pull my-weather-server --platform linux/amd64`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		serverName := args[0]

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

		fmt.Printf("Looking up server: %s\n", serverName)

		// Get server from database
		servers, err := database.GetServers()
		if err != nil {
			log.Fatalf("Failed to get servers: %v", err)
		}

		var targetServer *struct {
			name      string
			imageRef  string
			version   string
		}

		for _, server := range servers {
			if server.Name == serverName {
				// Parse server data to get Docker image reference
				imageRef, err := extractDockerImageRef(server.Data)
				if err != nil {
					log.Fatalf("Failed to extract Docker image reference: %v", err)
				}

				targetServer = &struct {
					name     string
					imageRef string
					version  string
				}{
					name:     server.Name,
					imageRef: imageRef,
					version:  server.Version,
				}
				break
			}
		}

		if targetServer == nil {
			log.Fatalf("Server not found: %s\nRun 'arctl list mcp' to see available servers", serverName)
		}

		fmt.Printf("✓ Found server: %s (version %s)\n", targetServer.name, targetServer.version)
		fmt.Printf("  Docker image: %s\n", targetServer.imageRef)

		// Pull the Docker image
		fmt.Println("\n🐳 Pulling Docker image...")
		if err := pullDockerImage(targetServer.imageRef, pullPlatform); err != nil {
			log.Fatalf("Failed to pull Docker image: %v", err)
		}

		fmt.Println("\n✅ MCP server pulled successfully!")
		fmt.Printf("\nTo install and run this server:\n")
		fmt.Printf("  arctl install mcp %s\n", serverName)
	},
}

func pullDockerImage(imageRef, platform string) error {
	args := []string{"pull"}
	
	if platform != "" {
		args = append(args, "--platform", platform)
	}
	
	args = append(args, imageRef)

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker pull failed: %w", err)
	}

	fmt.Printf("✓ Image pulled: %s\n", imageRef)
	return nil
}

func extractDockerImageRef(serverDataJSON string) (string, error) {
	// Parse the server data JSON to extract the Docker image reference
	// The image ref is in packages[0].identifier for Docker-based servers
	
	// Simple string parsing approach (could use json.Unmarshal for robustness)
	// Look for "identifier": "..." in the JSON
	
	start := -1
	searchStr := `"identifier"`
	for i := 0; i < len(serverDataJSON)-len(searchStr); i++ {
		if serverDataJSON[i:i+len(searchStr)] == searchStr {
			start = i + len(searchStr)
			break
		}
	}
	
	if start == -1 {
		return "", fmt.Errorf("could not find identifier in server data")
	}
	
	// Find the opening quote
	for i := start; i < len(serverDataJSON); i++ {
		if serverDataJSON[i] == '"' {
			start = i + 1
			break
		}
	}
	
	// Find the closing quote
	end := -1
	for i := start; i < len(serverDataJSON); i++ {
		if serverDataJSON[i] == '"' {
			end = i
			break
		}
	}
	
	if end == -1 {
		return "", fmt.Errorf("malformed identifier in server data")
	}
	
	imageRef := serverDataJSON[start:end]
	if imageRef == "" {
		return "", fmt.Errorf("empty identifier in server data")
	}
	
	return imageRef, nil
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.Flags().StringVarP(&pullPlatform, "platform", "p", "", "Target platform (e.g., linux/amd64)")
}

