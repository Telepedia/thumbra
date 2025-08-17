package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/telepedia/thumbra/config"
	"github.com/telepedia/thumbra/handlers"
	"github.com/telepedia/thumbra/storage"
)

func main() {
	cfg := config.Load()
	s3Client := storage.New(cfg.S3)

	r := mux.NewRouter()
	handlers.SetupRoutes(r, s3Client)

	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + cfg.Server.Port,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
	}

	log.Printf("Server listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
