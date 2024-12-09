package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"movieexample.com/rating/internal/controller/rating"
	"movieexample.com/rating/pkg/model"
)

// Handler is a struct that contains a reference to a rating.Controller.
// It is likely used to handle HTTP requests related to the rating functionality.
type Handler struct {
	ctrl *rating.Controller
}

// NewHandler creates a new instance of the Handler struct with the provided rating.Controller.
func NewHandler(ctrl *rating.Controller) *Handler {
	return &Handler{
		ctrl: ctrl,
	}
}

// Handle processes an HTTP request related to the rating functionality. It supports
// GET requests to retrieve the aggregate rating for a given record ID and type, and
// PUT requests to update the rating for a given record ID, type, and user ID.
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	// Handle the HTTP request here
	recorID := model.RecordID(r.FormValue("id"))
	if recorID == "" {
		http.Error(w, "Invalid record ID", http.StatusBadRequest)
		return
	}

	recordType := model.RecordType(r.FormValue("type"))
	if recordType == "" {
		http.Error(w, "Invalid record type", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodGet:
		v, err := h.ctrl.GetAggregateRating(r.Context(), recorID, recordType)
		if err != nil && errors.Is(err, rating.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if err := json.NewEncoder(w).Encode(v); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	case http.MethodPut:
		userID := model.UserID(r.FormValue("userId"))
		v, err := strconv.ParseFloat(r.FormValue("value"), 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := h.ctrl.PutRating(r.Context(), recorID, recordType, &model.Rating{
			UserID: userID,
			Value:  model.RatingValue(v),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
