package handlers

import (
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

func FetchCityClimate(w http.ResponseWriter, r *http.Request) {
    now := time.Now()
    unixTimestamp := now.Unix()
    unixTimestampRoundedToHour := (unixTimestamp / 3600) * 3600

    apiKey := os.Getenv("METEO_API_KEY")
    if apiKey == "" {
        log.Fatal("API_KEY environment variable is not set.")
    }

    // the response only contains temperature data, we have to wait for the full API acess to get more datapoints
    // the API also reports on only 50~ Sensors
    url := "https://www.meteoblue.com/de/products/cityclimate/getData?locationId=2657896&type=temperature&units=m&time=" + strconv.FormatInt(unixTimestampRoundedToHour, 10) + "&apikey=" + apiKey

    resp, err := http.Get(url)
    if err != nil {
        log.Printf("Error fetching city climate data: %v\n", err)
        http.Error(w, "Failed to fetch climate data", http.StatusInternalServerError)
        return
    }
    defer resp.Body.Close()

    w.Header().Set("Content-Type", "application/json")
    _, err = io.Copy(w, resp.Body)
    if err != nil {
        log.Printf("Error writing response: %v\n", err)
        http.Error(w, "Failed to write response", http.StatusInternalServerError)
    }
}

