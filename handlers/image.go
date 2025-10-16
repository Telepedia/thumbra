package handlers

import (
	"errors"
	"net/http"
	"strconv"

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

// Serve the original image back to the caller
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

// Actually return the image back to the caller (either by signaling that the browsers copy
// is fine to use), or by retrieving it from S3
func (h *ImageHandler) serveImage(w http.ResponseWriter, r *http.Request, req models.ImageRequest) {
	// validate that the request URL was actually valid
	err := utils.ValidateImageRequest(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// here we conditionally check the metadata such as etag, last modified
	metadata, err := h.imageService.GetImageMetadata(req)
	if err != nil {
		if err == services.ErrImageNotFound {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve image metadata", http.StatusInternalServerError)
		return
	}

	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		if ifNoneMatch == metadata.ETag || ifNoneMatch == "*" {
			if metadata.ETag != "" {
				w.Header().Set("ETag", metadata.ETag)
			}
			if !metadata.LastModified.IsZero() {
				w.Header().Set("Last-Modified", metadata.LastModified.UTC().Format(http.TimeFormat))
			}
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	if ifModSince := r.Header.Get("If-Modified-Since"); ifModSince != "" {
		if t, err := http.ParseTime(ifModSince); err == nil {
			if !metadata.LastModified.IsZero() && !metadata.LastModified.After(t) {
				if metadata.ETag != "" {
					w.Header().Set("ETag", metadata.ETag)
				}
				w.Header().Set("Last-Modified", metadata.LastModified.UTC().Format(http.TimeFormat))
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	// if we got here, we need to get the image from S3 and return it (either
	// expired or we fucked up!)
	obj, err := h.imageService.GetOriginalImage(req)
	if err != nil {
		if err == services.ErrImageNotFound {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve image", http.StatusInternalServerError)
		return
	}

	// set headers - note we don't need to set the cache-control as the middleware
	// handles that
	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(obj.Length, 10))

	if obj.ETag != "" {
		w.Header().Set("ETag", obj.ETag)
	}
	if obj.ContentDisposition != "" {
		w.Header().Set("Content-Disposition", obj.ContentDisposition)
	}
	if !obj.LastModified.IsZero() {
		w.Header().Set("Last-Modified", obj.LastModified.UTC().Format(http.TimeFormat))
	}

	_, _ = w.Write(obj.Data)
}

// Serve the original thumbnail back to the caller, generating it if it doesn't exist
func (h *ImageHandler) ServeThumbnail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	req := models.ThumbnailRequest{
		Wiki:     vars["wiki"],
		Hash1:    vars["hash1"],
		Hash2:    vars["hash2"],
		Filename: vars["filename"],
		Revision: vars["revision"],
		Width:    vars["width"],
	}

	h.serveThumbnail(w, r, req)
}

// Check if the requested thumbnail exists in S3, if so, return it. If it doesn't exist, attempt to scale
// it, store it in S3, and return it
func (h *ImageHandler) serveThumbnail(w http.ResponseWriter, r *http.Request, req models.ThumbnailRequest) {
	// validate that the request URL was actually valid
	err := utils.ValidateThumbnailRequest(req)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	metadata, err := h.imageService.GetThumbnailMetadata(req)
	if err != nil {
		if errors.Is(err, services.ErrImageNotFound) {
			// no thumbnail exists
			// @TODO: short circuit and generate the thumbnail when I can be bothered to implement that
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		if ifNoneMatch == metadata.ETag || ifNoneMatch == "*" {
			if metadata.ETag != "" {
				w.Header().Set("ETag", metadata.ETag)
			}
			if !metadata.LastModified.IsZero() {
				w.Header().Set("Last-Modified", metadata.LastModified.UTC().Format(http.TimeFormat))
			}
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	if ifModSince := r.Header.Get("If-Modified-Since"); ifModSince != "" {
		if t, err := http.ParseTime(ifModSince); err == nil {
			if !metadata.LastModified.IsZero() && !metadata.LastModified.After(t) {
				if metadata.ETag != "" {
					w.Header().Set("ETag", metadata.ETag)
				}
				w.Header().Set("Last-Modified", metadata.LastModified.UTC().Format(http.TimeFormat))
				w.WriteHeader(http.StatusNotModified)
				return
			}
		}
	}

	// not a 304, but the thumbnail exists so, lets return it!
	obj, err := h.imageService.GetThumbnail(req)

	if err != nil {
		if err == services.ErrImageNotFound {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve image", http.StatusInternalServerError)
		return
	}

	// set headers - note we don't need to set the cache-control as the middleware
	// handles that
	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(obj.Length, 10))

	if obj.ETag != "" {
		w.Header().Set("ETag", obj.ETag)
	}
	if obj.ContentDisposition != "" {
		w.Header().Set("Content-Disposition", obj.ContentDisposition)
	}
	if !obj.LastModified.IsZero() {
		w.Header().Set("Last-Modified", obj.LastModified.UTC().Format(http.TimeFormat))
	}

	_, _ = w.Write(obj.Data)

}
