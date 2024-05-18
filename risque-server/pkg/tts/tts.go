package tts

import (
	"fmt"
	"os/exec"
)

// TextToSpeech converts the given text to speech using eSpeak and saves it to a file.
func TextToSpeech(text, filepath string) error {
    // Command to save speech to a file in German language
    cmd := exec.Command("espeak", "-v", "en", text, "-w", filepath)
    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("failed to execute espeak command: %v", err)
    }
    return nil
}
