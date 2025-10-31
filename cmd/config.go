package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config <key> [value]",
	Short: "Get or set configuration values",
	Long: `Get or set arctl configuration values.

Available configuration keys:
  registry    Default Docker registry for push operations
              (e.g., docker.io/username, ghcr.io/username)

Examples:
  # Set default registry
  arctl config registry ghcr.io/myusername

  # Get current registry
  arctl config registry

  # View all config
  arctl config`,
	Args: cobra.RangeArgs(0, 2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// Show all config
			showAllConfig()
			return
		}

		key := args[0]
		
		if len(args) == 1 {
			// Get config value
			value, err := getConfig(key)
			if err != nil {
				log.Fatalf("Failed to get config: %v", err)
			}
			if value == "" {
				fmt.Printf("%s is not set\n", key)
			} else {
				fmt.Printf("%s = %s\n", key, value)
			}
			return
		}

		// Set config value
		value := args[1]
		if err := setConfig(key, value); err != nil {
			log.Fatalf("Failed to set config: %v", err)
		}
		fmt.Printf("✓ Set %s = %s\n", key, value)
	},
}

func getConfigDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".arctl")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create config directory: %w", err)
	}

	return configDir, nil
}

func getConfigPath() (string, error) {
	configDir, err := getConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config"), nil
}

func getConfig(key string) (string, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("failed to read config: %w", err)
	}

	// Parse simple key=value format
	lines := splitLines(string(data))
	for _, line := range lines {
		parts := splitKeyValue(line)
		if len(parts) == 2 && parts[0] == key {
			return parts[1], nil
		}
	}

	return "", nil
}

func setConfig(key, value string) error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	// Read existing config
	var lines []string
	data, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read config: %w", err)
	}
	if err == nil {
		lines = splitLines(string(data))
	}

	// Update or add the key
	found := false
	for i, line := range lines {
		parts := splitKeyValue(line)
		if len(parts) == 2 && parts[0] == key {
			lines[i] = fmt.Sprintf("%s=%s", key, value)
			found = true
			break
		}
	}

	if !found {
		lines = append(lines, fmt.Sprintf("%s=%s", key, value))
	}

	// Write back
	content := joinLines(lines)
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func showAllConfig() {
	configPath, err := getConfigPath()
	if err != nil {
		log.Fatalf("Failed to get config path: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No configuration set")
			fmt.Println("\nAvailable keys:")
			fmt.Println("  registry    Default Docker registry")
			return
		}
		log.Fatalf("Failed to read config: %v", err)
	}

	if len(data) == 0 {
		fmt.Println("No configuration set")
		return
	}

	fmt.Println("Configuration:")
	lines := splitLines(string(data))
	for _, line := range lines {
		if line != "" {
			fmt.Printf("  %s\n", line)
		}
	}
}

func splitLines(s string) []string {
	var lines []string
	var current string
	for _, c := range s {
		if c == '\n' {
			lines = append(lines, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitKeyValue(line string) []string {
	for i, c := range line {
		if c == '=' {
			return []string{line[:i], line[i+1:]}
		}
	}
	return []string{line}
}

func joinLines(lines []string) string {
	result := ""
	for i, line := range lines {
		result += line
		if i < len(lines)-1 {
			result += "\n"
		}
	}
	return result
}

// GetDefaultRegistry returns the default registry from config
func GetDefaultRegistry() string {
	registry, err := getConfig("registry")
	if err != nil {
		return ""
	}
	return registry
}

func init() {
	rootCmd.AddCommand(configCmd)
}

