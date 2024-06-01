package llm

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"server/pkg/models"
)

// GenerateSentence sends weather data to the Groq API and returns a human-readable sentence.
func GenerateSentence(data string) string {
    apiKey := os.Getenv("GROQ_API_KEY")
    if apiKey == "" {
        log.Fatal("GROQ_API_KEY environment variable is not set.")
    }

    url := "https://api.groq.com/openai/v1/chat/completions"
    payload := map[string]interface{}{
        "messages": []map[string]string{
            {"role": "user", "content": data},
        },
        "model": "llama3-70b-8192", // Ensure you are using the correct model
    }
    jsonData, err := json.Marshal(payload)
    if err != nil {
        log.Fatalf("Error encoding request data: %v", err)
    }

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        log.Fatalf("Error creating request: %v", err)
    }

    req.Header.Set("Authorization", "Bearer "+apiKey)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        log.Fatalf("Error sending request to Groq API: %v", err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatalf("Error reading response body: %v", err)
    }

    var apiResp models.APIResponse
    if err := json.Unmarshal(body, &apiResp); err != nil {
        log.Fatalf("Error decoding response from Groq API: %v", err)
    }

    if len(apiResp.Choices) > 0 && len(apiResp.Choices[0].Message.Content) > 0 {
        return apiResp.Choices[0].Message.Content
    }
    return "No sentence generated."
}
