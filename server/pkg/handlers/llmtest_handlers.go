package handlers

import (
	"log"
	"net/http"
	"server/pkg/llm"
)

func TestLLM(w http.ResponseWriter, r *http.Request) {

	// Use a simple, clear statement to see how the API responds
testSentence := "The temperature in Zurich is 4 degrees Celsius, and it is very cold."
interpretedText := llm.GenerateSentence(testSentence)
log.Println("Test API Output:", interpretedText)

    
}

