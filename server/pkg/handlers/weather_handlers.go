package handlers

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"server/pkg/llm"
	"server/pkg/models"
	"server/pkg/unrealspeech"
	"server/pkg/weatherdata"
	"time"
)

func FetchAndSpeakWeatherData(w http.ResponseWriter, r *http.Request) {
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

    var sum float64
    for _, feature := range cityClimateData.Features {
        sum += feature.Properties.Values
    }
    averageTemp := sum / float64(len(cityClimateData.Features))
    sentence := fmt.Sprintf("The current average temperature of the Sensor Grid is %.2f degrees Celsius. According to MeteoBlue, the temperature is %.2f degrees Celsius with a windspeed of %.2f meters per second. Please take all of this information and present it like a weatherman, in case of extreme temperature or windspeed, notify about the risk or give tips about how to avoid getting hurt by it. Please only mention this in case of actual risks, stay under 300 characters MAXIMUM!", averageTemp, firstTemperature, firstWindspeed)

    // Optionally enhance the sentence using your language model
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

    // Using predefined settings for the Unreal Speech API
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