package main

import (
	"fmt"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	"snippets.adelh.dev/app/internal/api"
	"snippets.adelh.dev/app/internal/config"
	"snippets.adelh.dev/app/internal/db"
	"snippets.adelh.dev/app/internal/encryption"
)

func main() {
	c, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	encryptionSvc, err := encryption.NewService(c.Enc.SystemKey)
	if err != nil {
		log.Fatal(err)
	}

	store, err := db.NewPostgresStore(c.DB)
	if err != nil {
		log.Fatal(err)
	}
	service := api.New(store, encryptionSvc)

	mux := http.NewServeMux()

	handler := api.HandlerFromMux(service, mux)

	addr := fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)

	srv := http.Server{
		Handler: handler,
		Addr:    addr,
	}

	fmt.Println("Server starting on ", addr)
	log.Fatal(srv.ListenAndServe())
}
