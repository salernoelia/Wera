package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"server/pkg/tts"
)

// RequestBody struct to hold the JSON request body
type RequestBody struct {
	Text string `json:"text"`
}

func TTSTest(w http.ResponseWriter, r *http.Request) {

	// Read the request body
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// Unmarshal the JSON request body
	var requestBody RequestBody
	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Convert text to speech
	err = tts.GoogleTextToSpeech(requestBody.Text, "output.wav")
	if err != nil {
		log.Fatalf("Failed to convert text to speech: %v", err)
	}

	log.Println("Audio content written to file 'output.wav'")
}
