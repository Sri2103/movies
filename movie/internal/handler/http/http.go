package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"movieexample.com/movie/internal/controller/movie"
)

// Handler is an HTTP handler that wraps a movie.Controller to handle HTTP requests.
type Handler struct {
	controller *movie.Controller
}

// New creates a new HTTP handler with the given movie controller.
func New(controller *movie.Controller) *Handler {
	return &Handler{
		controller: controller,
	}
}

// GetMoviedetails is an HTTP handler that retrieves movie details by the provided ID.
// If the movie is not found, it returns a 404 Not Found status.
// If there is an error retrieving the movie details, it returns a 500 Internal Server Error status.
// The movie details are encoded and written to the response writer.
func (h *Handler) GetMoviedetails(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	MovieDetails, err := h.controller.Get(r.Context(), id)
	if err != nil && errors.Is(err, movie.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(MovieDetails); err != nil {
		log.Printf("Response Encode err: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

}
