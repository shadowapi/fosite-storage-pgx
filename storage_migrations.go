package storage

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/jackc/tern/v2/migrate"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func (p *PgStorage) MigrateUp(ctx context.Context) error {
	m, err := migrate.NewMigrator(ctx, p.db.Conn(), p.migrationTableName)
	if err != nil {
		return fmt.Errorf("unable to create migrator: %v", err)
	}
	m.Data["TablesPrefix"] = p.tablesPrefix

	migrationsFS, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("unable to create sub filesystem: %v", err)
	}

	if err := m.LoadMigrations(migrationsFS); err != nil {
		return fmt.Errorf("unable to load migrations: %v", err)
	}

	return m.Migrate(ctx)
}
