package app

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/ldej/api-ldej-nl/internal/app/db"
	"github.com/ldej/api-ldej-nl/pkg/httpx"
)

type ThingResponse struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Value string `json:"value"`

	Updated time.Time `json:"updated"`
	Created time.Time `json:"created"`
}

// GetThing godoc
// @Summary Get a thing
// @Description get thing by uuid
// @ID get-thing-by-uuid
// @Tags Thing
// @Param uuid path string true "UUID"
// @Success 200 {object} ThingResponse
// @Failure 404,500 {object} httpx.ErrorResponse
// @Router /thing/{uuid} [get]
func (s *Server) GetThing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	uuid := chi.URLParam(r, "uuid")

	thing, err := s.db.GetThing(ctx, uuid)
	if err == db.ErrThingNotFound {
		httpx.AbortJSON(w, r, http.StatusNotFound, err)
		return
	}
	if err != nil {
		httpx.AbortJSON(w, r, http.StatusInternalServerError, err)
		return
	}
	httpx.JSON(w, r, thingToThingResponse(thing))
}

type CreateThing struct {
	Name  string `json:"name" validate:"required"`
	Value string `json:"value" validate:"required"`
}

// CreateThing godoc
// @Summary Create a thing
// @Description Create a thing
// @ID create-thing
// @Tags Thing
// @Param Body body CreateThing true "The body to create a thing"
// @Success 200 {object} ThingResponse
// @Failure 404,500 {object} httpx.ErrorResponse
// @Router /thing/new [post]
func (s *Server) CreateThing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var thingToCreate CreateThing
	err := s.parseJSON(r, &thingToCreate)
	if err != nil {
		httpx.AbortJSON(w, r, http.StatusBadRequest, err)
		return
	}

	createdThing, err := s.db.CreateThing(ctx, thingToCreate.Name, thingToCreate.Value)
	if err != nil {
		httpx.AbortJSON(w, r, http.StatusInternalServerError, err)
		return
	}

	httpx.JSON(w, r, thingToThingResponse(createdThing))
}

type UpdateThing struct {
	Value string `json:"value" validate:"required"`
}

// UpdateThing godoc
// @Summary Update a thing
// @Description Update a thing
// @ID update-thing
// @Tags Thing
// @Param uuid path string true "UUID"
// @Param Body body UpdateThing true "The body to update a thing"
// @Success 200 {object} ThingResponse
// @Failure 404,500 {object} httpx.ErrorResponse
// @Router /thing/{uuid} [put]
func (s *Server) UpdateThing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := chi.URLParam(r, "uuid")

	var thingToUpdate UpdateThing
	err := s.parseJSON(r, &thingToUpdate)
	if err != nil {
		httpx.AbortJSON(w, r, http.StatusBadRequest, err)
		return
	}

	updatedThing, err := s.db.UpdateThing(ctx, uuid, thingToUpdate.Value)
	if err == db.ErrThingNotFound {
		httpx.AbortJSON(w, r, http.StatusNotFound, err)
		return
	}
	if err != nil {
		httpx.AbortJSON(w, r, http.StatusInternalServerError, err)
		return
	}

	httpx.JSON(w, r, thingToThingResponse(updatedThing))
}

// DeleteThing godoc
// @Summary Delete a thing
// @Description Delete a thing
// @ID delete-thing
// @Tags Thing
// @Param uuid path string true "UUID"
// @Success 200 {object} ThingResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /thing/{uuid} [delete]
func (s *Server) DeleteThing(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	uuid := chi.URLParam(r, "uuid")

	err := s.db.DeleteThing(ctx, uuid)
	if err != nil {
		httpx.AbortJSON(w, r, http.StatusInternalServerError, err)
		return
	}
}

type ThingsResponse struct {
	Total  int             `json:"total"`
	Page   int             `json:"page"`
	Limit  int             `json:"limit"`
	Things []ThingResponse `json:"things"`
}

// ListThings godoc
// @Summary List things
// @Description List things
// @ID list-things
// @Tags Thing
// @Param page query int false "Page"
// @Param limit query int false "Limit (max 100)"
// @Success 200 {object} ThingsResponse
// @Failure 500 {object} httpx.ErrorResponse
// @Router /thing [get]
func (s *Server) ListThings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page := 1
	pageQueryParam := r.URL.Query().Get("page")
	if pageQueryParam != "" {
		result, err := strconv.Atoi(pageQueryParam)
		if err == nil && result > 0 {
			page = result
		}
	}

	limit := 10
	limitQueryParam := r.URL.Query().Get("limit")
	if limitQueryParam != "" {
		result, err := strconv.Atoi(limitQueryParam)
		if err == nil && result > 0 && result <= 100 {
			limit = result
		}
	}

	offset := 0
	if page > 1 {
		offset = (page - 1) * limit
	}

	things, count, err := s.db.GetThings(ctx, offset, limit)
	if err != nil {
		httpx.AbortJSON(w, r, http.StatusInternalServerError, err)
		return
	}

	thingsResponse := ThingsResponse{
		Page:   page,
		Limit:  limit,
		Total:  count,
		Things: []ThingResponse{},
	}
	for _, thing := range things {
		thingsResponse.Things = append(thingsResponse.Things, thingToThingResponse(thing))
	}
	httpx.JSON(w, r, thingsResponse)
}

func thingToThingResponse(thing db.Thing) ThingResponse {
	return ThingResponse{
		UUID:    thing.UUID,
		Name:    thing.Name,
		Value:   thing.Value,
		Updated: thing.Updated,
		Created: thing.Created,
	}
}
