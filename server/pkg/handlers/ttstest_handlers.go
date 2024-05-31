package handlers

import (
	"log"
	"net/http"
	"os"
	"server/pkg/tts"
)

func TTSTest(w http.ResponseWriter, r *http.Request) {

    // Print GOOGLE_APPLICATION_CREDENTIALS to verify it's set correctly
    log.Printf("GOOGLE_APPLICATION_CREDENTIALS: %s\n", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))

    // Convert text to speech
    err := tts.GoogleTextToSpeech("Hallo Welt", "output.mp3")
    if err != nil {
        log.Fatalf("Failed to convert text to speech: %v", err)
    }

    log.Println("Audio content written to file 'output.mp3'")
}
