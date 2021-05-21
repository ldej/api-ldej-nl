package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-playground/validator/v10"

	"github.com/ldej/api-ldej-nl/internal/app/db"
	"github.com/ldej/api-ldej-nl/pkg/log"
)

type Server struct {
	router   *chi.Mux
	log      *log.Logger
	db       db.Service
	validate *validator.Validate
	stopCh   chan os.Signal
}

func NewServer(logger *log.Logger, db db.Service) (*Server, error) {
	s := &Server{
		log:      logger,
		db:       db,
		validate: validator.New(),
		stopCh:   make(chan os.Signal, 1),
	}
	s.Routes()
	return s, nil
}

func (s *Server) Routes() {
	s.router = chi.NewRouter()
	s.router.Use(s.log.Tracer)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(60 * time.Second))

	s.router.Handle("/swagger/*", http.StripPrefix("/swagger", http.FileServer(http.Dir("swagger"))))

	s.router.Get("/thing", s.ListThings)
	s.router.Post("/thing/new", s.CreateThing)
	s.router.Get("/thing/{uuid}", s.GetThing)
	s.router.Put("/thing/{uuid}", s.UpdateThing)
	s.router.Delete("/thing/{uuid}", s.DeleteThing)
}

func (s *Server) ListenAndServe(addr string) {
	ctx := context.Background()

	signal.Notify(s.stopCh, syscall.SIGINT, syscall.SIGTERM)
	hs := &http.Server{Addr: addr, Handler: s.router}

	go func() {
		s.log.Info(ctx, fmt.Sprintf("Listening on: %s", addr))

		if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.log.Fatal(ctx, err)
		}
	}()

	<-s.stopCh
	s.log.Info(ctx, "Shutting down the server...")

	ctxTimeout, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	if err := hs.Shutdown(ctxTimeout); err != nil {
		s.log.Fatal(ctx, err)
	}
}

func (s *Server) Shutdown() {
	s.stopCh <- os.Interrupt
}

func (s *Server) parseJSON(r *http.Request, dst interface{}) error {
	err := json.NewDecoder(r.Body).Decode(dst)
	if err != nil {
		return errors.New("invalid json body")
	}
	if err := s.validate.Struct(dst); err != nil {
		return errors.New("invalid data")
	}
	return nil
}
