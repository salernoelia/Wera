package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
)

func FetchMeteoBlue(w http.ResponseWriter, r *http.Request) {

     apiKey := os.Getenv("METEO_API_KEY")
    if apiKey == "" {
        log.Fatal("API_KEY environment variable is not set.")
    }

    // URL is already prebuilt w/ Zürich coordinates
    url := "https://my.meteoblue.com/packages/basic-3h?apikey=" + apiKey + "&lat=47.3667&lon=8.55&asl=429&format=json" 

    resp, err := http.Get(url)
    if err != nil {
        log.Printf("Error fetching city climate data: %v\n", err)
        http.Error(w, "Failed to fetch climate data", http.StatusInternalServerError)
        return
    }
	// close body to avoid  pool error
    defer resp.Body.Close()

    w.Header().Set("Content-Type", "application/json")
    _, err = io.Copy(w, resp.Body)
    if err != nil {
        log.Printf("Error writing response: %v\n", err)
        http.Error(w, "Failed to write response", http.StatusInternalServerError)
    }
}
