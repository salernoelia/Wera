package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"risque-server/pkg/models"
	"risque-server/pkg/tts"
	"strconv"
	"time"
)

func FetchMeteoBlueData() (models.MeteoBlueData, error) {
    url := "https://my.meteoblue.com/packages/basic-3h?apikey=6MX8Tjra7uGLn2y9&lat=47.3667&lon=8.55&asl=429&format=json"
    resp, err := http.Get(url)
    if err != nil {
        return models.MeteoBlueData{}, err
    }
    defer resp.Body.Close()

    var meteoData models.MeteoBlueData
    if err := json.NewDecoder(resp.Body).Decode(&meteoData); err != nil {
        return models.MeteoBlueData{}, err
    }

 
    if len(meteoData.Data.Temperature) == 0 { // Example check
        return models.MeteoBlueData{}, errors.New("no temperature data received from MeteoBlue API")
    }

    return meteoData, nil
}


// FetchCityClimateData fetches data from CityClimate API
func FetchCityClimateData() (models.CityClimateData, error) {
    now := time.Now()
    unixTimestamp := now.Unix()
    unixTimestampRoundedToHour := (unixTimestamp / 3600) * 3600

    apiKey := "6MX8Tjra7uGLn2y9"
    url := "https://www.meteoblue.com/de/products/cityclimate/getData?locationId=2657896&type=temperature&units=m&time=" + strconv.FormatInt(unixTimestampRoundedToHour, 10) + "&apikey=" + apiKey

    resp, err := http.Get(url)
    if err != nil {
        return models.CityClimateData{}, err
    }
    defer resp.Body.Close()

    var cityClimateData models.CityClimateData
    if err := json.NewDecoder(resp.Body).Decode(&cityClimateData); err != nil {
        return models.CityClimateData{}, err
    }

    if len(cityClimateData.Features) == 0 { 
        return models.CityClimateData{}, errors.New("invalid data received from CityClimate API")
    }

    return cityClimateData, nil
}


func FetchAndSpeakWeatherData(w http.ResponseWriter, r *http.Request) {
    meteoData, err := FetchMeteoBlueData()
    if err != nil {
        log.Printf("Error fetching MeteoBlue data: %v\n", err)
        http.Error(w, "Failed to fetch MeteoBlue data", http.StatusInternalServerError)
        return
    }

    // Assuming that the arrays are not empty and data for the same index is synchronous
    if len(meteoData.Data.Temperature) == 0 || len(meteoData.Data.Windspeed) == 0 || len(meteoData.Data.PrecipitationProbability) == 0 {
        log.Println("Incomplete data received from MeteoBlue API")
        http.Error(w, "Incomplete data received", http.StatusInternalServerError)
        return
    }

    firstTemperature := meteoData.Data.Temperature[0]
    firstWindspeed := meteoData.Data.Windspeed[0]

    cityClimateData, err := FetchCityClimateData()
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
    sentence := fmt.Sprintf("The current average temperature of the Sensor Grid is %.2f degrees Celsius. According to MeteoBlue, the temperature is %.2f degrees Celsius with a windspeed of %.2f meters per second.", averageTemp, firstTemperature, firstWindspeed)
    log.Println(sentence)

    rand.Seed(time.Now().UnixNano())
    randomID := rand.Int()
    audioFileName := fmt.Sprintf("weather_%d.wav", randomID)
    filePath := filepath.Join("audio_files", audioFileName)

    err = tts.TextToSpeech(sentence, filePath)
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

    // Copy the file content to the response writer
    if _, err := io.Copy(w, file); err != nil {
        log.Printf("Error writing file to response: %v\n", err)
        http.Error(w, "Failed to send audio file", http.StatusInternalServerError)
    }
}
