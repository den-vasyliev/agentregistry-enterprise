package models

import "time"

// Registry represents a connected registry
type Registry struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Type      string    `json:"type"` // metadata only, not used functionally
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ServerDetail represents an MCP server from the registry
// Based on the MCP server.json schema
type ServerDetail struct {
	ID           int       `json:"id"`
	RegistryID   int       `json:"registry_id"`
	RegistryName string    `json:"registry_name"`
	Name         string    `json:"name"`        // e.g., "io.github.user/weather"
	Title        string    `json:"title"`       // Optional display name
	Description  string    `json:"description"` // Clear explanation of functionality
	Version      string    `json:"version"`     // Semantic version
	WebsiteURL   string    `json:"website_url"` // Optional homepage
	Installed    bool      `json:"installed"`
	Data         string    `json:"data"` // JSON blob of full server.json
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Skill represents a skill from the registry
type Skill struct {
	ID           int       `json:"id"`
	RegistryID   int       `json:"registry_id"`
	RegistryName string    `json:"registry_name"`
	Name         string    `json:"name"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Version      string    `json:"version"`
	Category     string    `json:"category"` // e.g., "data-processing", "communication", "automation"
	Installed    bool      `json:"installed"`
	Data         string    `json:"data"` // JSON blob
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Agent represents an AI agent from the registry
type Agent struct {
	ID          int       `json:"id"`
	RegistryID  int       `json:"registry_id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Model       string    `json:"model"`     // e.g., "gpt-4", "claude-3-opus"
	Specialty   string    `json:"specialty"` // e.g., "coding", "research", "customer-support"
	Installed   bool      `json:"installed"`
	Data        string    `json:"data"` // JSON blob
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Installation represents an installed resource (MCP server, skill, or agent)
type Installation struct {
	ID           int       `json:"id"`
	ResourceType string    `json:"resource_type"` // mcp, skill, agent
	ResourceID   int       `json:"resource_id"`
	ResourceName string    `json:"resource_name"`
	Version      string    `json:"version"`
	Config       string    `json:"config"` // JSON blob for configuration
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
