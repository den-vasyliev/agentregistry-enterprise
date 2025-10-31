package detector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCPServerType represents the type of MCP server
type MCPServerType string

const (
	TypeKMCP    MCPServerType = "kmcp"
	TypeNPM     MCPServerType = "npm"
	TypeUV      MCPServerType = "uv"
	TypeUnknown MCPServerType = "unknown"
)

// MCPServerInfo contains information about the detected MCP server
type MCPServerInfo struct {
	Type           MCPServerType
	Name           string
	Version        string
	Description    string
	Framework      string // For kmcp servers: fastmcp-python, etc.
	EntryPoint     string
	PackageManager string // npm, pnpm, yarn, uv
	HasDockerfile  bool
	RootDir        string
}

// DetectMCPServer analyzes a directory to determine MCP server type and configuration
func DetectMCPServer(dir string) (*MCPServerInfo, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	info := &MCPServerInfo{
		RootDir:       absDir,
		Type:          TypeUnknown,
		HasDockerfile: fileExists(filepath.Join(absDir, "Dockerfile")),
	}

	// Check for kmcp.yaml first (kmcp-based)
	kmcpYAMLPath := filepath.Join(absDir, "kmcp.yaml")
	if fileExists(kmcpYAMLPath) {
		if err := detectKMCPServer(info, kmcpYAMLPath); err != nil {
			return nil, err
		}
		info.Type = TypeKMCP
		return info, nil
	}

	// Check for package.json (npm-based)
	packageJSONPath := filepath.Join(absDir, "package.json")
	if fileExists(packageJSONPath) {
		if err := detectNPMServer(info, packageJSONPath); err != nil {
			return nil, err
		}
		info.Type = TypeNPM
		return info, nil
	}

	// Check for pyproject.toml (uv-based)
	pyprojectPath := filepath.Join(absDir, "pyproject.toml")
	if fileExists(pyprojectPath) {
		if err := detectUVServer(info, pyprojectPath); err != nil {
			return nil, err
		}
		info.Type = TypeUV
		return info, nil
	}

	// Check for requirements.txt (Python without uv)
	if fileExists(filepath.Join(absDir, "requirements.txt")) {
		info.Type = TypeUV
		info.PackageManager = "pip"
		// Try to find main entry point
		for _, candidate := range []string{"main.py", "server.py", "app.py", "__main__.py"} {
			if fileExists(filepath.Join(absDir, candidate)) {
				info.EntryPoint = candidate
				break
			}
		}
		if info.EntryPoint == "" {
			return nil, fmt.Errorf("could not find Python entry point (main.py, server.py, etc.)")
		}
		return info, nil
	}

	return nil, fmt.Errorf("could not detect MCP server type (no kmcp.yaml, package.json, pyproject.toml, or requirements.txt found)")
}

func detectKMCPServer(info *MCPServerInfo, kmcpYAMLPath string) error {
	data, err := os.ReadFile(kmcpYAMLPath)
	if err != nil {
		return fmt.Errorf("failed to read kmcp.yaml: %w", err)
	}

	content := string(data)

	// Parse YAML fields
	info.Name = extractYAMLValue(content, "name")
	info.Version = extractYAMLValue(content, "version")
	info.Description = extractYAMLValue(content, "description")
	info.Framework = extractYAMLValue(content, "framework")

	// kmcp servers should have a Dockerfile already
	if !info.HasDockerfile {
		return fmt.Errorf("kmcp server at %s is missing Dockerfile", info.RootDir)
	}

	// Package manager is determined by framework
	if contains(info.Framework, "python") {
		info.PackageManager = "python"
	} else if contains(info.Framework, "node") || contains(info.Framework, "typescript") {
		info.PackageManager = "npm"
	} else {
		info.PackageManager = "docker" // Generic for kmcp
	}

	return nil
}

func detectNPMServer(info *MCPServerInfo, packageJSONPath string) error {
	data, err := os.ReadFile(packageJSONPath)
	if err != nil {
		return fmt.Errorf("failed to read package.json: %w", err)
	}

	var pkg struct {
		Name    string            `json:"name"`
		Version string            `json:"version"`
		Main    string            `json:"main"`
		Bin     map[string]string `json:"bin"`
		Scripts map[string]string `json:"scripts"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return fmt.Errorf("failed to parse package.json: %w", err)
	}

	info.Name = pkg.Name
	info.Version = pkg.Version

	// Determine entry point
	if len(pkg.Bin) > 0 {
		// Use first bin entry
		for _, binPath := range pkg.Bin {
			info.EntryPoint = binPath
			break
		}
	} else if pkg.Main != "" {
		info.EntryPoint = pkg.Main
	} else {
		info.EntryPoint = "index.js"
	}

	// Detect package manager
	info.PackageManager = detectNodePackageManager(info.RootDir)

	return nil
}

func detectUVServer(info *MCPServerInfo, pyprojectPath string) error {
	// For now, we'll do basic detection
	// TODO: Parse pyproject.toml properly
	data, err := os.ReadFile(pyprojectPath)
	if err != nil {
		return fmt.Errorf("failed to read pyproject.toml: %w", err)
	}

	// Check if uv is being used
	if contains(string(data), "[tool.uv]") {
		info.PackageManager = "uv"
	} else {
		info.PackageManager = "pip"
	}

	// Try to extract name and version from pyproject.toml
	// This is a simple approach - ideally we'd use a TOML parser
	info.Name = extractTOMLValue(string(data), "name")
	info.Version = extractTOMLValue(string(data), "version")

	// Look for common Python entry points
	for _, candidate := range []string{"main.py", "server.py", "app.py", "__main__.py"} {
		if fileExists(filepath.Join(info.RootDir, candidate)) {
			info.EntryPoint = candidate
			break
		}
	}

	if info.EntryPoint == "" {
		return fmt.Errorf("could not find Python entry point")
	}

	return nil
}

func detectNodePackageManager(dir string) string {
	if fileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
		return "pnpm"
	}
	if fileExists(filepath.Join(dir, "yarn.lock")) {
		return "yarn"
	}
	if fileExists(filepath.Join(dir, "package-lock.json")) {
		return "npm"
	}
	return "npm" // default
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (len(s) >= len(substr)) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func extractYAMLValue(content, key string) string {
	// Basic YAML value extraction
	// Look for pattern: key: value
	lines := splitLines(content)
	for _, line := range lines {
		// Trim leading spaces
		trimmed := trimSpaces(line)

		// Check if line starts with the key
		if len(trimmed) > len(key)+1 && trimmed[:len(key)] == key && trimmed[len(key)] == ':' {
			// Extract value after colon
			value := trimmed[len(key)+1:]
			value = trimSpaces(value)

			// Remove quotes if present
			if len(value) > 0 && (value[0] == '"' || value[0] == '\'') {
				if len(value) >= 2 {
					value = value[1 : len(value)-1]
				}
			}

			return value
		}
	}
	return ""
}

func extractTOMLValue(content, key string) string {
	// Very basic TOML value extraction
	// Look for pattern: key = "value"
	lines := splitLines(content)
	for _, line := range lines {
		if contains(line, key+" =") {
			// Extract value between quotes
			start := -1
			end := -1
			inQuote := false
			for i, c := range line {
				if c == '"' || c == '\'' {
					if !inQuote {
						start = i + 1
						inQuote = true
					} else {
						end = i
						break
					}
				}
			}
			if start != -1 && end != -1 {
				return line[start:end]
			}
		}
	}
	return ""
}

func trimSpaces(s string) string {
	start := 0
	end := len(s)

	// Trim leading spaces
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}

	// Trim trailing spaces
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\r') {
		end--
	}

	if start >= end {
		return ""
	}

	return s[start:end]
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
