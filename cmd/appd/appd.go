package main

import (
	"context"
	"fmt"
	"os"

	"github.com/ldej/api-ldej-nl/internal/app"
	"github.com/ldej/api-ldej-nl/internal/app/db/datastoredb"
	"github.com/ldej/api-ldej-nl/pkg/log"
	_ "github.com/ldej/api-ldej-nl/swagger"
)

// @title api.ldej.nl
// @version 1.0
// @description A thing server
// @termsOfService http://swagger.io/terms/

// @contact.name Laurence de Jong
// @contact.url https://ldej.nl/
// @contact.email support@ldej.nl

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

func main() {
	projectID := "api-ldej-nl"
	ctx := context.Background()

	port := os.Getenv("PORT")
	if os.Getenv("PORT") == "" {
		port = "8080"
	}
	addr := fmt.Sprintf(":%s", port)

	logger := log.NewJSONLogger(os.Stderr, projectID, true)

	dbService, err := datastoredb.NewService(ctx, projectID)
	if err != nil {
		logger.Fatal(ctx, err)
	}

	server, err := app.NewServer(logger, dbService)
	if err != nil {
		logger.Fatal(ctx, err)
	}

	server.ListenAndServe(addr)
}
