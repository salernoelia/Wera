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
	"strings"
	"time"
)

func FetchAndSpeakWeatherBasedOnGPS(w http.ResponseWriter, r *http.Request) {
    var body models.RadioRequestBody
    err := json.NewDecoder(r.Body).Decode(&body)
    if err != nil {
        http.Error(w, "Invalid JSON input", http.StatusBadRequest)
        return
    }

    meteoData, meteoErr := weatherdata.FetchMeteoBlueData()
    if meteoErr != nil {
        log.Printf("Error fetching MeteoBlue data: %v\n", meteoErr)
    }

    cityClimateData, climateErr := weatherdata.FetchCityClimateData()
    if climateErr != nil {
        log.Printf("Error fetching CityClimate data: %v\n", climateErr)
    }

    closestSensor, distance := handlegps.FindClosestSensor(cityClimateData, body.Latitude, body.Longitude)
    if closestSensor == nil {
        log.Println("No close sensor found")
    }

    hotAreas, hotErr := weatherdata.FindHotAreas(cityClimateData)
    if hotErr != nil {
        log.Printf("No hot areas found: %v", hotErr)
    }

    var hotAreaNames []string
    for _, area := range hotAreas {
        hotAreaNames = append(hotAreaNames, area.Name)
    }

    var sentence string
    if closestSensor != nil {
        sentence = fmt.Sprintf("The closest sensor is %s, located %.2f km away. ", closestSensor.Name, distance)
    }
    if len(meteoData.Data.Temperature) > 0 {
        averageTemp := averageTemperature(cityClimateData)
        sentence += fmt.Sprintf("The current average temperature of the Sensor Grid is %.2f degrees Celsius. ", averageTemp)
        sentence += fmt.Sprintf("According to MeteoBlue, the temperature is %.2f degrees Celsius with a windspeed of %.2f meters. ", meteoData.Data.Temperature[0], meteoData.Data.Windspeed[0])
    }
    if len(hotAreaNames) > 0 {
        sentence += fmt.Sprintf("Be aware of hot areas like %s. ", strings.Join(hotAreaNames, ", "))
    }

    sentence += "Generate a few sentences like a weather speaker (dont claim to be one) nicely packed around this data, sounding very personalized, without actually mentioning any numbers."
    sentence += "Only say things that fit to the actual weather no hypotheticals."
    sentence += "Be friendly since you are talking to an elderly or non technical person, so do not mention any technicalities like sensors. If the temperatures are very high (over 30) please mention the risks of heat strokes and tell them to be cautious."
    sentence += "Please keep it short, maximum of 350 characters!"

    interpretedText := llm.GenerateSentence(sentence)
    log.Println("Original Sentence", sentence )
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

func averageTemperature(data models.CityClimateData) float64 {
    var sum float64
    for _, feature := range data.Features {
        sum += feature.Properties.Values
    }
    return sum / float64(len(data.Features))
}
