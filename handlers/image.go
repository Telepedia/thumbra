package handlers

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/telepedia/thumbra/models"
	"github.com/telepedia/thumbra/services"
	"github.com/telepedia/thumbra/storage"
	"github.com/telepedia/thumbra/utils"
)

type ImageHandler struct {
	s3Client     *storage.S3Client
	imageService *services.ImageService
}

func NewImageHandler(s3Client *storage.S3Client) *ImageHandler {
	return &ImageHandler{
		s3Client:     s3Client,
		imageService: services.NewImageService(s3Client),
	}
}

func (h *ImageHandler) ServeOriginal(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	req := models.ImageRequest{
		Wiki:     vars["wiki"],
		Hash1:    vars["hash1"],
		Hash2:    vars["hash2"],
		Filename: vars["filename"],
		Revision: vars["revision"],
	}

	h.serveImage(w, r, req)
}

func (h *ImageHandler) serveImage(w http.ResponseWriter, r *http.Request, req models.ImageRequest) {
	// validate that the request URL was actually valid
	err := utils.ValidateImageRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	imageData, contentType, err := h.imageService.GetOriginalImage(req)
	if err != nil {
		http.Error(w, "Failed to retrieve image", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(imageData)))
	w.Write(imageData)
}
