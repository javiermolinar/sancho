package db

import "fmt"

// migrate runs database migrations.
func (s *SQLite) migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS tasks (
			id              INTEGER PRIMARY KEY AUTOINCREMENT,
			description     TEXT NOT NULL,
			category        TEXT CHECK(category IN ('deep', 'shallow')),
			scheduled_date  DATE NOT NULL,
			scheduled_start TIME NOT NULL,
			scheduled_end   TIME NOT NULL,
			status          TEXT DEFAULT 'scheduled' CHECK(status IN ('scheduled', 'postponed', 'cancelled')),
			outcome         TEXT CHECK(outcome IN ('on_time', 'over', 'under')),
			postponed_from  INTEGER REFERENCES tasks(id),
			created_at      DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE INDEX IF NOT EXISTS idx_tasks_scheduled ON tasks(scheduled_date);
		CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
	`

	if _, err := s.db.Exec(query); err != nil {
		return fmt.Errorf("creating tasks table: %w", err)
	}

	return nil
}
