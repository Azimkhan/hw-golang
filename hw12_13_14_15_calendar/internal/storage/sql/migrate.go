package sqlstorage

import (
	"context"
	"fmt"
	"io/fs"
	"time"

	"github.com/Azimkhan/hw12_13_14_15_calendar/assets"
	"github.com/Azimkhan/hw12_13_14_15_calendar/internal/logger"
	"github.com/jackc/tern/v2/migrate"
)

const schemaVersionTable = "schema_version"

func MigrateDB(ctx context.Context, logg *logger.Logger, storage *Storage) error {
	migrator, err := migrate.NewMigrator(ctx, storage.conn, schemaVersionTable)
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
	migrator.OnStart = func(_ int32, name, direction, sql string) {
		logg.Info(
			fmt.Sprintf(
				"%s executing %s %s\n%s\n\n", time.Now().Format("2006-01-02 15:04:05"), name, direction, sql),
		)
	}
	err = migrator.Migrate(ctx)
	if err != nil {
		return fmt.Errorf("failed to migrate: %w", err)
	}
	return err
}
