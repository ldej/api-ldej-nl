// +build integration

package postgresdb

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/ldej/api-ldej-nl/internal/app/db"
	"github.com/ldej/api-ldej-nl/pkg/log"
	"github.com/ldej/api-ldej-nl/pkg/postgres"
	_ "github.com/ldej/api-ldej-nl/pkg/testing"
)

type Suite struct {
	suite.Suite
	db  db.Service
	ctx context.Context
}

func (s *Suite) SetupSuite() {
	s.ctx = context.Background()

	user := "postgres"
	pass := "mysecretpassword"
	host := "localhost"
	port := 5432
	dbName := "integration"

	logger := log.NewJSONLogger(os.Stderr, "", true)

	err := postgres.CreateDB(s.ctx, host, port, user, pass, dbName, true)
	s.NoError(err)

	err = postgres.ApplyMigrations(
		s.ctx,
		logger,
		fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", user, pass, host, port, dbName),
		"migrations",
	)
	s.NoError(err)

	s.db, err = NewService(s.ctx, "localhost", 5432, user, pass, dbName)
	s.NoError(err)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) TestThing() {
	thing, err := s.db.CreateThing(s.ctx, "name", "value")
	s.NoError(err)
	s.Equal("name", thing.Name)

	retrievedThing, err := s.db.GetThing(s.ctx, thing.UUID)
	s.NoError(err)
	s.Equal(thing.UUID, retrievedThing.UUID)

	things, count, err := s.db.GetThings(s.ctx, 0, 10)
	s.NoError(err)
	s.Equal(count, len(things))
	s.Equal(1, count)

	updatedThing, err := s.db.UpdateThing(s.ctx, thing.UUID, "updated")
	s.NoError(err)
	s.Equal("updated", updatedThing.Value)

	err = s.db.DeleteThing(s.ctx, thing.UUID)
	s.NoError(err)

	_, err = s.db.GetThing(s.ctx, thing.UUID)
	s.Equal(db.ErrThingNotFound, err)
}

func (s *Suite) TestThingNotFound() {
	_, err := s.db.UpdateThing(s.ctx, "does-not-exist", "value")
	s.Equal(db.ErrThingNotFound, err)
}
