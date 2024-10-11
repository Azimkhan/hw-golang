package sqlstorage

import (
	"context"
	"fmt"
	"github.com/Azimkhan/hw12_13_14_15_calendar/assets"
	"github.com/jackc/tern/v2/migrate"
	"io/fs"
)

const schemaVersionTable = "schema_version"

func MigrateDB(ctx context.Context, storage *Storage, callBack func(_ int32, name, direction, sql string)) error {
	migrator, err := migrate.NewMigrator(ctx, storage.Conn, schemaVersionTable)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	dir, err := fs.Sub(assets.Migrations, "migrations")
	if err != nil {
		return fmt.Errorf("failed to get migrations dir: %w", err)
	}
	err = migrator.LoadMigrations(dir)
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}
	migrator.OnStart = callBack
	err = migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}
	return err
}
