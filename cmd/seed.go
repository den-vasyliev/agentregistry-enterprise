package cmd

import (
	"fmt"
	"log"

	"github.com/agentregistry-dev/agentregistry/internal/database"
	"github.com/spf13/cobra"
)

var seedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Seed the database with mock data",
	Long:  `Populate the database with sample registries, servers, skills, and agents for testing.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Initialize database
		if err := database.Initialize(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer func() {
			if err := database.Close(); err != nil {
				log.Printf("Warning: Failed to close database: %v", err)
			}
		}()

		fmt.Println("Seeding database with mock data...")

		// Create mock registry
		fmt.Println("\n📦 Creating mock registry...")
		if err := database.AddRegistry("Mock AI Registry", "https://registry.example.com", "public"); err != nil {
			if contains(err.Error(), "UNIQUE constraint failed") {
				fmt.Println("  ⚠️  Mock registry already exists")
			} else {
				log.Fatalf("Failed to add mock registry: %v", err)
			}
		} else {
			fmt.Println("  ✓ Created 'Mock AI Registry'")
		}

		// Get registry ID
		registries, err := database.GetRegistries()
		if err != nil {
			log.Fatalf("Failed to get registries: %v", err)
		}

		var mockRegistryID int
		for _, reg := range registries {
			if reg.Name == "Mock AI Registry" {
				mockRegistryID = reg.ID
				break
			}
		}

		if mockRegistryID == 0 {
			log.Fatal("Could not find Mock AI Registry")
		}

		// Seed MCP Servers
		fmt.Println("\n🖥️  Seeding MCP servers...")
		servers := getMockServers()
		for _, server := range servers {
			if err := database.AddOrUpdateServer(
				mockRegistryID,
				server["name"].(string),
				server["title"].(string),
				server["description"].(string),
				server["version"].(string),
				server["website_url"].(string),
				server["data"].(string),
			); err != nil {
				log.Printf("  ⚠️  Failed to add server %s: %v", server["name"], err)
			} else {
				fmt.Printf("  ✓ Added %s\n", server["title"])
			}
		}

		// Seed Skills
		fmt.Println("\n⚡ Seeding skills...")
		skills := getMockSkills()
		for _, skill := range skills {
			if err := database.AddOrUpdateSkill(
				mockRegistryID,
				skill["name"].(string),
				skill["title"].(string),
				skill["description"].(string),
				skill["version"].(string),
				skill["category"].(string),
				skill["data"].(string),
			); err != nil {
				log.Printf("  ⚠️  Failed to add skill %s: %v", skill["name"], err)
			} else {
				fmt.Printf("  ✓ Added %s\n", skill["title"])
			}
		}

		// Seed Agents
		fmt.Println("\n🤖 Seeding agents...")
		agents := getMockAgents()
		for _, agent := range agents {
			if err := database.AddOrUpdateAgent(
				mockRegistryID,
				agent["name"].(string),
				agent["title"].(string),
				agent["description"].(string),
				agent["version"].(string),
				agent["model"].(string),
				agent["specialty"].(string),
				agent["data"].(string),
			); err != nil {
				log.Printf("  ⚠️  Failed to add agent %s: %v", agent["name"], err)
			} else {
				fmt.Printf("  ✓ Added %s\n", agent["title"])
			}
		}

		fmt.Println("\n✅ Database seeded successfully!")
		fmt.Println("\nYou can now:")
		fmt.Println("  - Run 'arctl ui' to view in the web interface")
		fmt.Println("  - Run 'arctl list mcp' to see MCP servers")
		fmt.Println("  - Run 'arctl list skill' to see skills")
		fmt.Println("  - Run 'arctl list agent' to see agents")
	},
}

func getMockServers() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "io.example/weather-api",
			"title":       "Weather API Server",
			"description": "Real-time weather data and forecasts using OpenWeatherMap API",
			"version":     "1.2.0",
			"website_url": "https://example.com/weather",
			"data": `{
				"name": "io.example/weather-api",
				"description": "Real-time weather data and forecasts",
				"version": "1.2.0",
				"packages": [{"identifier": "weather-api", "registryType": "npm"}]
			}`,
		},
		{
			"name":        "io.example/database-connector",
			"title":       "Database Connector",
			"description": "Universal database connector for PostgreSQL, MySQL, and MongoDB",
			"version":     "2.0.1",
			"website_url": "https://example.com/database",
			"data": `{
				"name": "io.example/database-connector",
				"description": "Universal database connector",
				"version": "2.0.1",
				"packages": [{"identifier": "database-connector", "registryType": "npm"}]
			}`,
		},
		{
			"name":        "io.example/file-system",
			"title":       "File System Operations",
			"description": "Secure file system access with sandboxing and permission controls",
			"version":     "1.5.0",
			"website_url": "https://example.com/filesystem",
			"data": `{
				"name": "io.example/file-system",
				"description": "Secure file system operations",
				"version": "1.5.0",
				"packages": [{"identifier": "file-system", "registryType": "npm"}]
			}`,
		},
	}
}

func getMockSkills() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "skill.example/data-analyzer",
			"title":       "Data Analyzer",
			"description": "Advanced data analysis and visualization capabilities with support for CSV, JSON, and Excel files",
			"version":     "1.0.0",
			"category":    "data-processing",
			"data": `{
				"name": "skill.example/data-analyzer",
				"capabilities": ["csv-parsing", "data-visualization", "statistical-analysis"],
				"version": "1.0.0"
			}`,
		},
		{
			"name":        "skill.example/email-composer",
			"title":       "Email Composer",
			"description": "Intelligent email drafting with tone adjustment, grammar checking, and template management",
			"version":     "2.1.0",
			"category":    "communication",
			"data": `{
				"name": "skill.example/email-composer",
				"capabilities": ["tone-adjustment", "grammar-check", "templates"],
				"version": "2.1.0"
			}`,
		},
		{
			"name":        "skill.example/task-scheduler",
			"title":       "Task Scheduler",
			"description": "Automated task scheduling and workflow management with calendar integration",
			"version":     "1.3.2",
			"category":    "automation",
			"data": `{
				"name": "skill.example/task-scheduler",
				"capabilities": ["scheduling", "reminders", "calendar-sync"],
				"version": "1.3.2"
			}`,
		},
		{
			"name":        "skill.example/code-reviewer",
			"title":       "Code Reviewer",
			"description": "Automated code review with security scanning, best practices checking, and refactoring suggestions",
			"version":     "3.0.0",
			"category":    "development",
			"data": `{
				"name": "skill.example/code-reviewer",
				"capabilities": ["security-scan", "best-practices", "refactoring"],
				"version": "3.0.0"
			}`,
		},
		{
			"name":        "skill.example/content-summarizer",
			"title":       "Content Summarizer",
			"description": "Intelligent content summarization for articles, documents, and research papers with key point extraction",
			"version":     "1.5.1",
			"category":    "content",
			"data": `{
				"name": "skill.example/content-summarizer",
				"capabilities": ["summarization", "key-points", "multi-language"],
				"version": "1.5.1"
			}`,
		},
	}
}

func getMockAgents() []map[string]interface{} {
	return []map[string]interface{}{
		{
			"name":        "agent.example/code-assistant",
			"title":       "Code Assistant",
			"description": "AI coding assistant specialized in full-stack development with expertise in React, Python, and Node.js",
			"version":     "1.0.0",
			"model":       "gpt-4",
			"specialty":   "coding",
			"data": `{
				"name": "agent.example/code-assistant",
				"capabilities": ["code-generation", "debugging", "documentation"],
				"languages": ["javascript", "python", "typescript"],
				"version": "1.0.0"
			}`,
		},
		{
			"name":        "agent.example/research-assistant",
			"title":       "Research Assistant",
			"description": "Academic research assistant with access to scholarly databases and citation management",
			"version":     "2.0.0",
			"model":       "claude-3-opus",
			"specialty":   "research",
			"data": `{
				"name": "agent.example/research-assistant",
				"capabilities": ["literature-review", "citation-management", "data-analysis"],
				"version": "2.0.0"
			}`,
		},
		{
			"name":        "agent.example/customer-support",
			"title":       "Customer Support Agent",
			"description": "24/7 customer support agent with knowledge base integration and escalation management",
			"version":     "1.5.0",
			"model":       "gpt-4-turbo",
			"specialty":   "customer-support",
			"data": `{
				"name": "agent.example/customer-support",
				"capabilities": ["ticket-management", "kb-search", "sentiment-analysis"],
				"version": "1.5.0"
			}`,
		},
		{
			"name":        "agent.example/data-scientist",
			"title":       "Data Science Agent",
			"description": "Data science specialist for statistical analysis, machine learning, and predictive modeling",
			"version":     "1.2.0",
			"model":       "gpt-4",
			"specialty":   "data-science",
			"data": `{
				"name": "agent.example/data-scientist",
				"capabilities": ["statistical-analysis", "ml-models", "visualization"],
				"version": "1.2.0"
			}`,
		},
		{
			"name":        "agent.example/content-creator",
			"title":       "Content Creator",
			"description": "Creative writing agent for blog posts, social media, and marketing content with SEO optimization",
			"version":     "2.1.0",
			"model":       "claude-3-sonnet",
			"specialty":   "content-creation",
			"data": `{
				"name": "agent.example/content-creator",
				"capabilities": ["blog-writing", "social-media", "seo-optimization"],
				"version": "2.1.0"
			}`,
		},
		{
			"name":        "agent.example/devops-engineer",
			"title":       "DevOps Engineer",
			"description": "DevOps automation agent for CI/CD pipelines, infrastructure as code, and deployment management",
			"version":     "1.0.0",
			"model":       "gpt-4",
			"specialty":   "devops",
			"data": `{
				"name": "agent.example/devops-engineer",
				"capabilities": ["cicd", "infrastructure", "monitoring"],
				"version": "1.0.0"
			}`,
		},
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (len(s) >= len(substr)) && containsHelper(s, substr)
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(seedCmd)
}
