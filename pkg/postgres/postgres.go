package postgres

import (
	"context"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // golint: required for importing migration source files

	"github.com/ldej/api-ldej-nl/pkg/log"
)

func ApplyMigrations(ctx context.Context, logger *log.Logger, uri string, migrationsFolder string) error {
	m, err := migrate.New(
		"file://"+migrationsFolder,
		uri,
	)
	if err != nil {
		return err
	}

	err = m.Up()
	if err == migrate.ErrNoChange {
		logger.Info(ctx, "Already up to date")
	} else if err != nil {
		return err
	} else {
		logger.Info(ctx, "Successfully applied migrations")
	}
	return nil
}
