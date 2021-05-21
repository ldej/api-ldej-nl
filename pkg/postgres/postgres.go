package postgres

import (
	"context"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file" // golint: required for importing migration source files
	"github.com/jmoiron/sqlx"

	"github.com/ldej/api-ldej-nl/pkg/log"
)

func CreateDB(ctx context.Context, host string, port int, user string, pass string, name string, drop bool) error {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s sslmode=disable", host, port, user, pass)

	pg, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return err
	}

	if drop {
		_, err = pg.ExecContext(ctx, `DROP DATABASE IF EXISTS `+name)
		if err != nil {
			return err
		}
	}
	_, err = pg.ExecContext(ctx, `CREATE DATABASE `+name)
	return err
}

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
