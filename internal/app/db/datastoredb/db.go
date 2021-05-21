package datastoredb

import (
	"context"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/google/uuid"

	"github.com/ldej/api-ldej-nl/internal/app/db"
)

var _ db.Service = (*service)(nil)

const thingKind = "thing"

type service struct {
	datastoreClient *datastore.Client
}

func NewService(ctx context.Context, projectID string) (db.Service, error) {
	datastoreClient, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return &service{datastoreClient: datastoreClient}, nil
}

func (s *service) GetThing(ctx context.Context, uuid string) (db.Thing, error) {
	var thing db.Thing
	key := datastore.NameKey(thingKind, uuid, nil)
	err := s.datastoreClient.Get(ctx, key, &thing)
	if err == datastore.ErrNoSuchEntity {
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

	key := datastore.NameKey(thingKind, thing.UUID, nil)
	_, err := s.datastoreClient.Put(ctx, key, &thing)
	if err != nil {
		return db.Thing{}, err
	}
	return thing, nil
}

func (s *service) UpdateThing(ctx context.Context, uuid string, value string) (db.Thing, error) {
	now := time.Now().UTC()

	var thing db.Thing
	key := datastore.NameKey(thingKind, uuid, nil)
	err := s.datastoreClient.Get(ctx, key, &thing)
	if err == datastore.ErrNoSuchEntity {
		return db.Thing{}, db.ErrThingNotFound
	}
	if err != nil {
		return db.Thing{}, err
	}

	thing.Value = value
	thing.Updated = now

	_, err = s.datastoreClient.Put(ctx, key, &thing)
	if err != nil {
		return db.Thing{}, err
	}
	return thing, nil
}

func (s *service) DeleteThing(ctx context.Context, uuid string) error {
	key := datastore.NameKey(thingKind, uuid, nil)
	return s.datastoreClient.Delete(ctx, key)
}

func (s *service) GetThings(ctx context.Context, offset int, limit int) ([]db.Thing, int, error) {
	var things []db.Thing
	query := datastore.NewQuery(thingKind).Offset(offset).Limit(limit)
	_, err := s.datastoreClient.GetAll(ctx, query, &things)
	if err != nil {
		return nil, 0, err
	}
	count, err := s.datastoreClient.Count(ctx, datastore.NewQuery(thingKind))
	if err != nil {
		return nil, 0, err
	}
	return things, count, nil
}
