package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/telepedia/thumbra/models"
	"github.com/telepedia/thumbra/services"
	"github.com/telepedia/thumbra/storage"
	"github.com/telepedia/thumbra/utils"
)

// For now, much of this is duplicated from ImageHandler for development
// we can eventually clean up and combine what we can
type ThumbnailHandler struct {
	s3Client     *storage.S3Client
	imageService *services.ImageService
}

func NewThumbnailHandler(s3Client *storage.S3Client) *ThumbnailHandler {
	return &ThumbnailHandler{
		s3Client:     s3Client,
		imageService: services.NewImageService(s3Client),
	}
}

func (h *ThumbnailHandler) ServeThumbnail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	req := models.ThumbnailRequest{
		Wiki:     vars["wiki"],
		Hash1:    vars["hash1"],
		Hash2:    vars["hash2"],
		Filename: vars["filename"],
		Revision: vars["revision"],
		Width:    vars["width"],
	}

	h.serve(w, r, req)
}

func (h *ThumbnailHandler) serve(w http.ResponseWriter, r *http.Request, req models.ThumbnailRequest) {
	err := utils.ValidateThumbnailRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	
}
