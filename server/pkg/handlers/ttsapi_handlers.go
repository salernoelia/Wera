package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"server/pkg/models"
	"server/pkg/tts"
)

func SpeakText(w http.ResponseWriter, r *http.Request) {
    var ttsRequest models.TTSRequest
    err := json.NewDecoder(r.Body).Decode(&ttsRequest)
    if err != nil {
        http.Error(w, "Invalid JSON input", http.StatusBadRequest)
        return
    }

    if ttsRequest.Text == "" {
        http.Error(w, "Text field is required", http.StatusBadRequest)
        return
    }

    // Define the file path where the audio will be saved
    filePath := filepath.Join("audio_files", "output.wav")

    // Ensure the directory exists
    err = tts.TextToSpeech(ttsRequest.Text, filePath)
    if err != nil {
        log.Printf("Error converting text to speech: %v", err)
        http.Error(w, "Failed to convert text to speech", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    response := map[string]string{"message": "Text spoken successfully", "file": filePath}
    json.NewEncoder(w).Encode(response)
}
