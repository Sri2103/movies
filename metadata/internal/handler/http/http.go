package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"movieexample.com/metadata/internal/controller/metadata"
	"movieexample.com/metadata/pkg/model"
)

type Handler struct {
	// ctrl is a pointer to a metadata.Controller instance, which is used to handle
	// metadata-related operations in the HTTP handler.
	ctrl *metadata.Controller
}

// New creates a new instance of the Handler struct, which is used to handle
// HTTP-related metadata operations. The ctrl parameter is a pointer to a
// metadata.Controller instance, which is used to handle the underlying
// metadata-related logic.
func New(ctrl *metadata.Controller) *Handler {
	return &Handler{
		ctrl: ctrl,
	}
}

// GetMetadata is an HTTP handler that retrieves metadata for a given ID. If the ID is empty,
// it returns a 400 Bad Request response. If the metadata is not found, it returns a 404 Not Found
// response. If there is an error encoding the metadata, it returns a 500 Internal Server Error
// response. Otherwise, it encodes the metadata as JSON and writes it to the response.
func (h *Handler) GetMetadata(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	m, err := h.ctrl.Get(ctx, id)
	if err != nil && errors.Is(err, metadata.ErrNotFound) {
		log.Printf("Repository got err: %v", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if err := json.NewEncoder(w).Encode(m); err != nil {
		log.Printf("Failed to encode metadata: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}

// PutMetadata is an HTTP handler that updates the metadata for a given ID. It decodes the request body
// into a metadata.Metadata struct, and then calls the Put method on the metadata.Controller to update
// the metadata. If there is an error decoding the request body, it returns a 400 Bad Request response.
// If there is an error updating the metadata, it returns a 500 Internal Server Error response.
func (h *Handler) PutMetadata(w http.ResponseWriter, r *http.Request) {
	var m *model.Metadata
	if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
		log.Printf("Failed to decode metadata: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	if err := h.ctrl.Put(ctx, m); err != nil {
		log.Printf("Failed to put metadata: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
