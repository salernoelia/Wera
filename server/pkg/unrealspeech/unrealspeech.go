package unrealspeech

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"server/pkg/models"
)



func GenerateSpeech(req models.SpeechRequest, filePath string) error {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return fmt.Errorf("error encoding request data: %w", err)
    }

    client := &http.Client{}
    request, err := http.NewRequest("POST", "https://api.v6.unrealspeech.com/stream", bytes.NewReader(jsonData))
    if err != nil {
        return fmt.Errorf("error creating request: %w", err)
    }

    request.Header.Set("Authorization", "Bearer " + os.Getenv("UNREAL_SPEECH_API_KEY"))
    request.Header.Set("Content-Type", "application/json")

    resp, err := client.Do(request)
    if err != nil {
        return fmt.Errorf("error sending request to Unreal Speech API: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("Unreal Speech API returned non-OK status: %s, body: %s", resp.Status, string(bodyBytes))
    }

    outputFile, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("error creating audio file: %w", err)
    }
    defer outputFile.Close()

    _, err = io.Copy(outputFile, resp.Body)
    if err != nil {
        return fmt.Errorf("error saving audio file: %w", err)
    }

    return nil
}
