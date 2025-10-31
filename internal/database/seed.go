package database

import "fmt"

// AddOrUpdateSkill adds or updates a skill in the database
func AddOrUpdateSkill(registryID int, name, title, description, version, category, data string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := DB.Exec(`
		INSERT INTO skills (registry_id, name, title, description, version, category, data, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(registry_id, name, version) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			category = excluded.category,
			data = excluded.data,
			updated_at = CURRENT_TIMESTAMP
	`, registryID, name, title, description, version, category, data)
	if err != nil {
		return fmt.Errorf("failed to add/update skill: %w", err)
	}

	return nil
}

// AddOrUpdateAgent adds or updates an agent in the database
func AddOrUpdateAgent(registryID int, name, title, description, version, model, specialty, data string) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	_, err := DB.Exec(`
		INSERT INTO agents (registry_id, name, title, description, version, model, specialty, data, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(registry_id, name, version) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			model = excluded.model,
			specialty = excluded.specialty,
			data = excluded.data,
			updated_at = CURRENT_TIMESTAMP
	`, registryID, name, title, description, version, model, specialty, data)
	if err != nil {
		return fmt.Errorf("failed to add/update agent: %w", err)
	}

	return nil
}
