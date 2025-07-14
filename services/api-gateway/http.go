package main

import (
	"encoding/json"
	"log"
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


	log.Println("Received request to preview trip")

	writeJSON(w,http.StatusCreated,"ok")

}