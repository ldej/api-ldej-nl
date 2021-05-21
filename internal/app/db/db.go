package db

import (
	"context"
	"errors"
	"time"
)

type Service interface {
	GetThing(ctx context.Context, uuid string) (Thing, error)
	CreateThing(ctx context.Context, name string, value string) (Thing, error)
	UpdateThing(ctx context.Context, uuid string, value string) (Thing, error)
	DeleteThing(ctx context.Context, uuid string) error
	GetThings(ctx context.Context, offset int, limit int) ([]Thing, int, error)
}

type Thing struct {
	UUID  string `db:"uuid"`
	Name  string `db:"name"`
	Value string `db:"value"`

	Updated time.Time `db:"updated"`
	Created time.Time `db:"created"`
}

var (
	ErrThingNotFound = errors.New("thing not found")
)
