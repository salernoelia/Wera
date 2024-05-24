package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"server/pkg/handlegps"
	"server/pkg/llm"
	"server/pkg/models"
	"server/pkg/unrealspeech"
	"server/pkg/weatherdata"
	"time"
)




func FetchAndSpeakWeatherBasedOnGPS(w http.ResponseWriter, r *http.Request) {
    var body models.RadioRequestBody
    err := json.NewDecoder(r.Body).Decode(&body)
    if err != nil {
        http.Error(w, "Invalid JSON input", http.StatusBadRequest)
        return
    }

    meteoData, err := weatherdata.FetchMeteoBlueData()
    if err != nil {
        log.Printf("Error fetching MeteoBlue data: %v\n", err)
        http.Error(w, "Failed to fetch MeteoBlue data", http.StatusInternalServerError)
        return
    }

    if len(meteoData.Data.Temperature) == 0 || len(meteoData.Data.Windspeed) == 0 || len(meteoData.Data.PrecipitationProbability) == 0 {
        log.Println("Incomplete data received from MeteoBlue API")
        http.Error(w, "Incomplete data received", http.StatusInternalServerError)
        return
    }

    firstTemperature := meteoData.Data.Temperature[0]
    firstWindspeed := meteoData.Data.Windspeed[0]

    cityClimateData, err := weatherdata.FetchCityClimateData()
    if err != nil {
        log.Printf("Error fetching CityClimate data: %v\n", err)
        http.Error(w, "Failed to fetch CityClimate data", http.StatusInternalServerError)
        return
    }

    // Find the closest sensor
    closestSensor, distance := handlegps.FindClosestSensor(cityClimateData, body.Latitude, body.Longitude)
    if closestSensor == nil {
        http.Error(w, "No close sensor found", http.StatusInternalServerError)
        return
    }

    // Check if there are hot areas
    hotAreas, err := weatherdata.FindHotAreas(cityClimateData)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }

    // Location names of hot areas
    var hotAreaNames []string
    for _, area := range hotAreas {
        hotAreaNames = append(hotAreaNames, area.Name)
    }

    



    var sum float64
    for _, feature := range cityClimateData.Features {
        sum += feature.Properties.Values
    }
    averageTemp := sum / float64(len(cityClimateData.Features))
    fmt.Printf("Closest Sensor: %s, Distance: %.2f km\n, MeteoBlue Temperature: %.2f, Windspeed: %.2f\n, Sensor Temperature: %.2f\n", closestSensor.Name, distance, firstTemperature, firstWindspeed, averageTemp)
    sentence := fmt.Sprintf("The closest sensor is %s, located %.2f km away. The current average temperature of the Sensor Grid is %.2f degrees Celsius. According to MeteoBlue, the temperature is %.2f degrees Celsius with a windspeed of %.2f meters per second. Generate a few sentences like a weather speaker (dont claim to be one) nicely packed around this data without actually mentioning any numbers. Please keep it short, maximum of 250 characters! Be friendly since you are talking to an elderly.", closestSensor.Name, distance, averageTemp, firstTemperature, firstWindspeed)

    interpretedText := llm.GenerateSentence(sentence)
    log.Println("Interpreted Text: ", interpretedText)

    rand.Seed(time.Now().UnixNano())
    randomID := rand.Int()
    audioFileName := fmt.Sprintf("weather_%d.wav", randomID)
    audioDir := "audio_files"
    filePath := filepath.Join(audioDir, audioFileName)

    if err := os.MkdirAll(audioDir, os.ModePerm); err != nil {
        log.Printf("Error creating directory: %v\n", err)
        http.Error(w, "Failed to create directory for audio file", http.StatusInternalServerError)
        return
    }

    err = unrealspeech.GenerateSpeech(models.SpeechRequest{
        Text:    interpretedText,
        VoiceId: "Scarlett",
        Bitrate: "64k",
        Speed:   "0",
        Pitch:   "1",
        Codec:   "libmp3lame",
    }, filePath)

    if err != nil {
        log.Printf("Error converting text to speech: %v\n", err)
        http.Error(w, "Failed to convert text to speech", http.StatusInternalServerError)
        return
    }

    file, err := os.Open(filePath)
    if err != nil {
        log.Printf("Error opening file: %v\n", err)
        http.Error(w, "Failed to open audio file", http.StatusInternalServerError)
        return
    }
    defer file.Close()

    w.Header().Set("Content-Type", "audio/wav")

    if _, err := io.Copy(w, file); err != nil {
        log.Printf("Error writing file to response: %v\n", err)
        http.Error(w, "Failed to send audio file", http.StatusInternalServerError)
    }
}

