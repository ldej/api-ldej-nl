package postgresdb

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/ldej/api-ldej-nl/internal/app/db"
)

var _ db.Service = (*service)(nil)

type service struct {
	pg *sqlx.DB
}

func NewService(ctx context.Context, host string, port int, user string, pass string, name string) (db.Service, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name)
	pg, err := sqlx.Connect("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}
	return &service{pg: pg}, nil
}

func (s *service) GetThing(ctx context.Context, uuid string) (db.Thing, error) {
	var thing db.Thing
	err := s.pg.GetContext(
		ctx, &thing, `SELECT * FROM things WHERE uuid = $1`, uuid)
	if err == sql.ErrNoRows {
		return db.Thing{}, db.ErrThingNotFound
	}
	if err != nil {
		return db.Thing{}, err
	}
	return thing, nil
}

func (s *service) CreateThing(ctx context.Context, name string, value string) (db.Thing, error) {
	now := time.Now().UTC()
	thing := db.Thing{
		UUID:    uuid.New().String(),
		Name:    name,
		Value:   value,
		Updated: now,
		Created: now,
	}
	_, err := s.pg.NamedExecContext(
		ctx,
		`INSERT INTO things (uuid, name, value, updated, created) 
		    VALUES (:uuid, :name, :value, :updated, :created)`,
		thing,
	)
	if err != nil {
		return db.Thing{}, err
	}
	return thing, nil
}

func (s *service) UpdateThing(ctx context.Context, uuid string, value string) (db.Thing, error) {
	_, err := s.pg.ExecContext(
		ctx,
		`UPDATE things SET value = $1 WHERE uuid = $2`,
		value,
		uuid,
	)
	if err != nil {
		return db.Thing{}, err
	}
	return s.GetThing(ctx, uuid)
}

func (s *service) DeleteThing(ctx context.Context, uuid string) error {
	_, err := s.pg.ExecContext(
		ctx,
		`DELETE FROM things WHERE uuid = $1`,
		uuid,
	)
	return err
}

func (s *service) GetThings(ctx context.Context, offset int, limit int) ([]db.Thing, int, error) {
	var things []db.Thing
	err := s.pg.SelectContext(
		ctx,
		&things,
		`SELECT * FROM things OFFSET $1 LIMIT $2`,
		offset,
		limit,
	)
	if err != nil {
		return nil, 0, err
	}
	row := s.pg.QueryRowContext(
		ctx,
		`SELECT COUNT(*) as count FROM things`,
	)
	var count int
	err = row.Scan(&count)
	if err != nil {
		return nil, 0, err
	}
	return things, count, nil
}
