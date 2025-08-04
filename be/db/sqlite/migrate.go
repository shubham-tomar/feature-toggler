package sqlite

import "database/sql"

func Migrate(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			name TEXT,
			email TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			name TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS project_users (
			id TEXT PRIMARY KEY,
			user_id TEXT,
			project_id TEXT,
			role TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP
		);`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}
