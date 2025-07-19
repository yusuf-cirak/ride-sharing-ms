package http

import (
	"encoding/json"
	"net/http"
	"ride-sharing/services/trip-service/internal/domain"
	"ride-sharing/shared/types"
)


type HttpHandler struct {
	Service domain.TripService
}


type previewTripRequest struct {
	UserID string `json:"userId"`
	Pickup types.Coordinate `json:"pickup"`
	Destination types.Coordinate `json:"destination"`
}

func (h *HttpHandler) HandleTripReview(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	var reqBody previewTripRequest;
	if err:= json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}


	if reqBody.UserID=="" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	t, err := h.Service.GetRoute(ctx, &reqBody.Pickup, &reqBody.Destination)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w,http.StatusCreated, t)

}


func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}