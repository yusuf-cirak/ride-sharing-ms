package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

func handleTripReview(w http.ResponseWriter, r *http.Request) {

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

	jsonBody, err := json.Marshal(reqBody)

	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		return
	}
	
	reader := bytes.NewReader(jsonBody) // creating a new reader buffer because r.Body is already closed

	response, err := http.Post("http://trip-service:8083/trip/preview", "application/json", reader)
	
	if err != nil {
		http.Error(w, "Failed to connect to trip service", http.StatusInternalServerError)
		return
	}

	defer response.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(response.StatusCode)
	
	if _, err := io.Copy(w, response.Body); err != nil {
		// Log error if needed
		return
	}

}