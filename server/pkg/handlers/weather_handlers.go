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
	"strings"
	"time"
)

func FetchAndSpeakWeatherData(w http.ResponseWriter, r *http.Request) {
    meteoData, err := weatherdata.FetchMeteoBlueData()
    if err != nil {
        log.Printf("Error fetching MeteoBlue data: %v\n", err)
        http.Error(w, "Failed to fetch MeteoBlue data", http.StatusInternalServerError)
        return
    }

    if len(meteoData.Data1H.Temperature) == 0 || len(meteoData.Data1H.Windspeed) == 0 || len(meteoData.Data1H.PrecipitationProbability) == 0 {
        log.Println("Incomplete data received from MeteoBlue API")
        http.Error(w, "Incomplete data received", http.StatusInternalServerError)
        return
    }

    // current time and date
    currentTime := time.Now()


    // find current time spot in data_1H.time, format is 2024-06-02 06:00
    var currentTimeSpot int
    for i, timeSpot := range meteoData.Data1H.Time {
        if currentTime.Format("2006-01-02 15:04") == timeSpot {
            currentTimeSpot = i
            fmt.Println("Current time spot: ", currentTimeSpot)
            break
    }
}

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

    var sentence string

    if len(cityClimateData.Features) > 0 {
        averageTemp := averageTemperature(cityClimateData)
        sentence += fmt.Sprintf("The current average temperature of the Sensor Grid is %.2f degrees Celsius. ", averageTemp)
    }

    if len(meteoData.Data1H.Temperature) > 0 && len(meteoData.Data1H.Windspeed) > 0 && len(meteoData.Data1H.PrecipitationProbability) > 0{
        sentence += fmt.Sprintf("According to MeteoBlue, the current temperature is %.2f degrees Celsius with a windspeed of %.2f meters. ", meteoData.Data1H.Temperature[currentTimeSpot], meteoData.Data1H.Windspeed[currentTimeSpot])
        sentence += fmt.Sprintf("The relative humidity is %d percent. ", meteoData.Data1H.RelativeHumidity[currentTimeSpot])

        calculatedNext1HTemp, tempErr := temperatureNext1H(meteoData.Data1H.Temperature)

        if tempErr != nil {
            log.Printf("Error calculating next 3 hour temperature: %v", tempErr)
        } else {
             sentence+= fmt.Sprintf("The average temperature of the next thee hours is %.2f degrees Celsius. ", calculatedNext1HTemp)
        }

        peakTemp, timeOfPeakTemp := peakMeteoTemperature(meteoData)
        sentence += fmt.Sprintf("The peak temperature of the day is %.2f degrees Celsius at %s. ", peakTemp, timeOfPeakTemp)

        if peakTemp > 30 {
            sentence += "It will be above 30 degrees Celsius today. Be cautious of heat strokes."
        }

        

        peakWindspeed := peakMeteoWindspeed(meteoData)

        if peakWindspeed > 10 {
            sentence += fmt.Sprintf("The peak windspeed of the day is %.2f meters per second.  ", peakWindspeed)
        } else {
            sentence += "The windspeed is not expected to exceed 10 meters per second."
        }

        windy := willItBeWindy(meteoData)
        if len(windy) > 0 {
            sentence += fmt.Sprintf("It will be windy at %s. ", strings.Join(windy, ", "))
        }

        willRain := willItRain(meteoData)

        if len(willRain) > 0 {
            sentence += fmt.Sprintf("It will rain at %s. ", strings.Join(willRain, ", "))
        }

        willSnow := willItSnow(meteoData)
        if len(willSnow) > 0 {
            sentence += fmt.Sprintf("It will snow at %s. ", strings.Join(willSnow, ", "))
        }

        willFog := willItBeFoggy(meteoData)
        if len(willFog) > 0 {
            sentence += fmt.Sprintf("It will be foggy at %s. ", strings.Join(willFog, ", "))
        }

        willWind := willItBeWindy(meteoData)
        if len(willWind) > 0 {
            sentence += fmt.Sprintf("It will be windy at %s. ", strings.Join(willWind, ", "))
        }

        highUVIndex := willHaveHighUVIndex(meteoData)
        if len(highUVIndex) > 0 {
            sentence += fmt.Sprintf("There will be a high UV index at %s. ", strings.Join(highUVIndex, ", "))
        }

    }

    if len(meteoData.Data1H.PrecipitationProbability) > 0 {
        sentence += fmt.Sprintf("The precipitation probability is %d percent.", meteoData.Data1H.PrecipitationProbability[currentTimeSpot])
    }


    sentence += fmt.Sprintf("The current time is %s and the date is %s. ", currentTime.Format("15:04"), currentTime.Format("2006-01-02"))

    sentence += "Generate a few sentences like a nice friend (dont claim to be one) packed around the data, sounding very personalized, without actually mentioning any numbers, except for the time and date"
    sentence += "If there is a high windspeed or UV Index, mention it and give hints to prevent sunburn or getting hurt in case of strong winds."
    sentence += "Only say things that fit to the actual weather no hypotheticals. So dont give useless advice like 'wear a jacket' if its 30 degrees outside. Or 'stay in the shade' if its raining."
    sentence += "Be friendly since you are talking to an elderly or non technical person, so do not mention any technicalities like sensors. If the temperatures are very high (over 30) please mention the risks of heat strokes and tell them to be cautious."
    sentence += "Please keep it short, maximum of 400 characters! Your name is Wera."

  
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

