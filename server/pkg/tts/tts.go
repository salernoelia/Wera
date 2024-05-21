package tts

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func TextToSpeech(text, filePath string) error {
	apiKey := os.Getenv("TTS_API_KEY")
    if apiKey == "" {
        log.Fatal("TTS_API_KEY environment variable is not set.")
    }
	language := "en-us"
	voice := "Mike"
	codec := "WAV"
	

	url := fmt.Sprintf(
		"http://api.voicerss.org/?key=%s&hl=%s&v=%s&c=%s&src=%s",
		apiKey, language, voice, codec, strings.ReplaceAll(text, " ", "+"),
	)

	fmt.Println(url)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making request to Voice RSS API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error response from Voice RSS API: %s", resp.Status)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating audio file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("error saving audio file: %w", err)
	}

	return nil
}
