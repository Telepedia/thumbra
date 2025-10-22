package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/telepedia/thumbra/models"
	"github.com/telepedia/thumbra/public"
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

var placeholderImage = &models.ImageResponse{
	ContentType: "image/webp",
	Data:        public.PlaceholderData,
	Length:      int64(len(public.PlaceholderData)),
}

func writePlaceholderResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", placeholderImage.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(placeholderImage.Length, 10))
	// cache these for an hour
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusNotFound)
	_, _ = w.Write(placeholderImage.Data)
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
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// here we conditionally check the metadata such as etag, last modified
	metadata, err := h.imageService.GetImageMetadata(req)
	if err != nil {
		if err == services.ErrImageNotFound {
			log.Printf("image not found: %s/%s", req.Wiki, req.Filename)
			writePlaceholderResponse(w)
			return
		}
		log.Printf("failed to retrieve image metadata: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "An erorr occurred, please try again later.")
		return
	}

	if checkConditionalGet(w, r, metadata) {
		return
	}

	// if we got here, we need to get the image from S3 and return it (either
	// expired or we fucked up!)
	obj, err := h.imageService.GetOriginalImage(req)
	if err != nil {
		if err == services.ErrImageNotFound {
			log.Printf("image not found: %s/%s", req.Wiki, req.Filename)
			writePlaceholderResponse(w)
			return
		}
		log.Printf("failed to retrieve image: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "An erorr occurred, please try again later.")
		return
	}

	writeS3ObjectResponse(w, obj)
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

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(req.Filename), "."))
	if !utils.SupportedThumbFormats[ext] {
		// this is a pass through format, we need to return the original,
		// since we cannot thumbnail it
		// @TODO: maybe instead we move this to the thumb generation bit?
		model := models.ImageRequest{
			Wiki:     req.Wiki,
			Hash1:    req.Hash1,
			Hash2:    req.Hash2,
			Filename: req.Filename,
			Revision: req.Revision,
		}
		h.serveImage(w, r, model)
		return
	}

	// we can thumbnail this type of file, so generate the thumbnail
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
			// fetch the original image to generate the thumbnail
			model := models.ImageRequest{
				Wiki:     req.Wiki,
				Hash1:    req.Hash1,
				Hash2:    req.Hash2,
				Filename: req.Filename,
				Revision: req.Revision,
			}

			obj, err := h.imageService.GetOriginalImage(model)
			if err != nil {
				if err == services.ErrImageNotFound {
					log.Printf("image not found: %s/%s", req.Wiki, req.Filename)
					writePlaceholderResponse(w)
					return
				}
				log.Printf("Failed to retrieve original image during thumbnail process: %v", err)
				writeJSONError(w, http.StatusInternalServerError, "An erorr occurred, please try again later.")
				return
			}

			// here we pass the original image to the thumbnail generator, and the requested model
			// this function saves the thumbnail to the temp dir and returns the path. It is this functions
			// responsibility to upload the thumbnail to S3
			path, err := h.imageService.ThumbnailImage(req, obj)

			if err != nil {
				log.Printf("Failed to generate thumbnail: %v", err)
				writeJSONError(w, http.StatusInternalServerError, "An erorr occurred generating the thumbnail, please try again later.")
				return
			}

			// delete temp file when this func finishes
			// @TODO: this is a rough draft - we need to handle this better and split out a lot of
			// this into utility functions, but also remove a lot of the duplicated code
			defer os.Remove(path)

			// Upload the thumbnail to S3
			err = h.imageService.UploadThumbnail(req, path)
			if err != nil {
				log.Printf("Failed to upload thumbnail to S3: %v", err)
				writeJSONError(w, http.StatusInternalServerError, "An erorr occurred, please try again later.")
				return
			}

			// serve the thumbnail we just created
			thumbObj, err := h.imageService.GetThumbnail(req)
			if err != nil {
				log.Printf("Failed to retrieve thumbnail after upload: %v", err)
				writeJSONError(w, http.StatusInternalServerError, "An erorr occurred, please try again later.")
				return
			}

			// Set headers and return the thumbnail
			writeS3ObjectResponse(w, thumbObj)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if checkConditionalGet(w, r, metadata) {
		return
	}

	// not a 304, but the thumbnail exists so, lets return it!
	obj, err := h.imageService.GetThumbnail(req)

	if err != nil {
		if err == services.ErrImageNotFound {
			log.Printf("thumbnail should exist but was not found: %s/%s", req.Wiki, req.Filename)
			writePlaceholderResponse(w)
			return
		}
		log.Printf("failed to retrieve thumbnail: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "An erorr occurred, please try again later.")
		return
	}

	writeS3ObjectResponse(w, obj)

}

// Check whether we can send a 304 not modified response to the caller
// instead of returning the full object from S3
func checkConditionalGet(w http.ResponseWriter, r *http.Request, metadata *models.ImageResponse) bool {
	if metadata == nil {
		return false
	}

	if ifNoneMatch := r.Header.Get("If-None-Match"); ifNoneMatch != "" {
		if ifNoneMatch == metadata.ETag || ifNoneMatch == "*" {
			if metadata.ETag != "" {
				w.Header().Set("ETag", metadata.ETag)
			}
			if !metadata.LastModified.IsZero() {
				w.Header().Set("Last-Modified", metadata.LastModified.UTC().Format(http.TimeFormat))
			}
			w.Header().Set("Cache-Control", "public, max-age=31536000")
			w.WriteHeader(http.StatusNotModified)
			return true
		}
	}

	if ifModSince := r.Header.Get("If-Modified-Since"); ifModSince != "" {
		if t, err := http.ParseTime(ifModSince); err == nil {
			if !metadata.LastModified.IsZero() && !metadata.LastModified.After(t) {
				if metadata.ETag != "" {
					w.Header().Set("ETag", metadata.ETag)
				}
				w.Header().Set("Last-Modified", metadata.LastModified.UTC().Format(http.TimeFormat))
				w.Header().Set("Cache-Control", "public, max-age=31536000")
				w.WriteHeader(http.StatusNotModified)
				return true
			}
		}
	}

	return false
}

// Utility function to write common headers for S3 object responses
func writeS3ObjectResponse(w http.ResponseWriter, obj *models.ImageResponse) {
	if obj == nil {
		return
	}
	// set headers
	w.Header().Set("Content-Type", obj.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(obj.Length, 10))
	w.Header().Set("Cache-Control", "public, max-age=31536000")

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

// Utility function to write JSON error responses
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := map[string]string{
		"error": message,
	}
	_ = json.NewEncoder(w).Encode(resp)
}
