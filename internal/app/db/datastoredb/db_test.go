// +build integration

package datastoredb

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/suite"

	"github.com/ldej/api-ldej-nl/internal/app/db"
)

type Suite struct {
	suite.Suite
	db  db.Service
	ctx context.Context
}

func (s *Suite) SetupSuite() {
	s.ctx = context.Background()

	emulatorHost := os.Getenv("DATASTORE_EMULATOR_HOST")
	if emulatorHost == "" {
		// os.Setenv("DATASTORE_EMULATOR_HOST", "localhost:8081")
		s.T().Skip("No datastore emulator available")
	}

	resp, err := resty.New().R().Post(fmt.Sprintf("http://%s/reset", emulatorHost))
	s.NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode())

	s.db, err = NewService(s.ctx, "test")
	s.NoError(err)
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(Suite))
}

func (s *Suite) TestThing() {
	thing, err := s.db.CreateThing(s.ctx, "name", "value")
	s.NoError(err)

	retrievedThing, err := s.db.GetThing(s.ctx, thing.UUID)
	s.NoError(err)
	s.Equal("value", retrievedThing.Value)
	s.Equal("name", retrievedThing.Name)

	_, err = s.db.UpdateThing(s.ctx, thing.UUID, "updated")
	s.NoError(err)

	retrievedThing, err = s.db.GetThing(s.ctx, thing.UUID)
	s.NoError(err)
	s.Equal("updated", retrievedThing.Value)
	s.Equal("name", retrievedThing.Name)

	retrievedThings, count, err := s.db.GetThings(s.ctx, 0, 10)
	s.NoError(err)
	s.Equal(count, len(retrievedThings))
	s.Equal(1, count)

	err = s.db.DeleteThing(s.ctx, thing.UUID)
	s.NoError(err)

	_, err = s.db.GetThing(s.ctx, thing.UUID)
	s.Equal(db.ErrThingNotFound, err)
}

func (s *Suite) TestThingNotFound() {
	_, err := s.db.GetThing(s.ctx, "does-not-exist")
	s.Equal(db.ErrThingNotFound, err)

	_, err = s.db.UpdateThing(s.ctx, "does-not-exist", "updated")
	s.Equal(db.ErrThingNotFound, err)
}
