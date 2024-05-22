package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/pkg/unrealspeech"
)

func SpeakText(w http.ResponseWriter, r *http.Request) {
    var request unrealspeech.SpeechRequest
    err := json.NewDecoder(r.Body).Decode(&request)
    if err != nil {
        http.Error(w, "Invalid JSON input", http.StatusBadRequest)
        return
    }

    filePath := "output.mp3"  // Define path dynamically as needed
    err = unrealspeech.GenerateSpeech(request, filePath)
    if err != nil {
        log.Printf("Error generating speech: %v\n", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"message": "Speech generated successfully", "file": filePath})
}
