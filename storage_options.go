package storage

// WithTablesPrefix sets the prefix for all tables in the database.
func WithTablesPrefix(prefix string) option {
	return func(p *PgStorage) {
		p.tablesPrefix = prefix
	}
}

// WithMigrationTableName sets the name of the table used to store migration information.
// The default value is "public.auth_fosite_migrations".
// It's highly recommended to use a schema-qualified table name.
func WithMigrationTableName(name string) option {
	return func(p *PgStorage) {
		p.migrationTableName = name
	}
}
