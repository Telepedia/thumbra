package handlers

import (
	"github.com/gorilla/mux"
	"github.com/telepedia/thumbra/storage"
)

func SetupRoutes(r *mux.Router, s3Client *storage.S3Client) {
	imageHandler := NewImageHandler(s3Client)

	r.HandleFunc("/{wiki}/{hash1}/{hash2}/{filename}/revision/{revision}",
		imageHandler.ServeOriginal).Methods("GET")
}
