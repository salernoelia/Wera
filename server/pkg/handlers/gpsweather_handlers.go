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
	"server/pkg/tts"
	"server/pkg/weatherdata"
	"strings"
	"time"
)

func FetchAndSpeakWeatherBasedOnGPS(w http.ResponseWriter, r *http.Request) {
    var body models.RadioRequestBodyGPS
    err := json.NewDecoder(r.Body).Decode(&body)
    if err != nil {
        http.Error(w, "Invalid JSON input", http.StatusBadRequest)
        return
    }


    meteoData, meteoErr := weatherdata.FetchMeteoBlueData()
    if meteoErr != nil {
        log.Printf("Error fetching MeteoBlue data: %v\n", meteoErr)
    }

    // Switzerland is typically in CET or CEST, so let's assume CEST for now (+2 UTC)
    location, err := time.LoadLocation("Europe/Zurich")
    if err != nil {
        log.Printf("Error loading location 'Europe/Zurich': %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    currentTime := time.Now().In(location)

    // time rounded to hour
    currentTime = currentTime.Round(time.Hour)

    currentTimeSpot := -1 // Set to -1 to indicate "not found"

    // Find the closest time spot in data_1H.time
    for i, timeSpot := range meteoData.Data1H.Time {
        parsedTime, err := time.ParseInLocation("2006-01-02 15:04", timeSpot, location)
        if err != nil {
            log.Printf("Error parsing time '%s': %v", timeSpot, err)
            continue
        }

        // Check for an exact match
        if currentTime.Format("2006-01-02 15:04") == parsedTime.Format("2006-01-02 15:04") {
            currentTimeSpot = i
            break
        }
    }

    if currentTimeSpot == -1 {
        fmt.Fprint(w, "Current time spot not found. ")
    } else {
        fmt.Fprintf(w, "Current time spot index: %d. ", currentTimeSpot)
        fmt.Fprintf(w, "Current time spot: %s. ", meteoData.Data1H.Time[currentTimeSpot])
    }

    fmt.Println("Current time spot index:", currentTimeSpot)
    fmt.Fprint(w, "Server time (CET/CEST): ", currentTime.Format("2006-01-02 15:04"))

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

    fmt.Println(body.Language)

    

    sentence := "The name of the user is " + body.DeviceID + ". "
    sentence += "Your name is wera and you are a good friend of the user, are experienced in weather and want to help stay informed in a formal way, don't be overly excited or positive. "
    sentence += "Only in case of extreme weather conditions, like heat (28 degrees or above), you should give advice to the user and remind them of preventive measures. "
    sentence += "An example would be if there is a high windspeed or UV Index or if it is snowing or raining. "
    sentence += "Only say things that fit to the actual weather no hypotheticals. So dont give useless advice like 'wear a jacket' if its 30 degrees outside. Or 'stay in the shade' if its raining. "
    sentence += "Since you are talking to a non technical person, do not mention any technical words like sensors, percipitation or numbers and data like the celsius. "
    sentence += "Do not exceed 700 characters in your resonse, and formulate in a way like its being spoken. "
    sentence += "The current time is " + currentTime.Format("15:04") + " and the date is " + currentTime.Format("2006-01-02") + "but you only mention daytimes like 'morning' or 'afternoon'. "
    sentence += "Do not mention any numbers "
    sentence += "Data you have acess to, to form your weather report:"


    if closestSensor != nil {
        sentence += fmt.Sprintf("The closest City Temperature sensor is at %s, located %.2f km away. It reports a temperature of %.2f, mention this by saying something like 'around your house it is ...' ", closestSensor.Name, distance, closestSensor.Values)
    }

    if len(hotAreaNames) > 0 {
        sentence += fmt.Sprintf("Be especially aware of hot areas like %s. ", strings.Join(hotAreaNames, ", "))
    }

    if len(cityClimateData.Features) > 0 {
        averageTemp := weatherdata.AverageTemperature(cityClimateData)
        sentence += fmt.Sprintf("At the current time, the average of all City Temperature sensors is %.2f degrees Celsius. ", averageTemp)
    }

    if len(meteoData.Data1H.Temperature) > 0 && len(meteoData.Data1H.Windspeed) > 0 && len(meteoData.Data1H.PrecipitationProbability) > 0{
        sentence += fmt.Sprintf("The current temperature in Zürich, according to the meteo service, is %.2f degrees Celsius with a windspeed of %.2f meters. ", meteoData.Data1H.Temperature[currentTimeSpot], meteoData.Data1H.Windspeed[currentTimeSpot])
        sentence += fmt.Sprintf("The relative humidity is %d percent. ", meteoData.Data1H.RelativeHumidity[currentTimeSpot])

        TemperatureNext3H, calc3HError := weatherdata.TemperatureNext3H(meteoData.Data1H.Temperature)
        if calc3HError != nil {
            log.Printf("Error calculating next 6 hour temperature: %v", calc3HError)
        } else {
            sentence += fmt.Sprintf("The average temperature of the next three hours is %.2f degrees Celsius, ", TemperatureNext3H)
        }

        TemperatureNext6H, calc6HError := weatherdata.TemperatureNext6H(meteoData.Data1H.Temperature)

        if calc6HError != nil {
            log.Printf("Error calculating next 3 hour temperature: %v", calc6HError)
        } else {
             sentence+= fmt.Sprintf("and over the next six hours %.2f degrees Celsius. ", TemperatureNext6H)
        }

        peakTemp, timeOfPeakTemp := weatherdata.PeakMeteoTemperature(meteoData)
        sentence += fmt.Sprintf("The peak temperature of the day is %.2f degrees Celsius at %s. ", peakTemp, timeOfPeakTemp)


        peakWindspeed := weatherdata.PeakMeteoWindspeed(meteoData)
        sentence += fmt.Sprintf("The peak windspeed of the day is %.2f meters per second.  ", peakWindspeed)


        windy := weatherdata.WillItBeWindy(meteoData)
        if len(windy) > 0 {
            sentence += fmt.Sprintf("It will be windy at %s. ", strings.Join(windy, ", "))
        }

        willRain := weatherdata.WillItRain(meteoData)

        if len(willRain) > 0 {
            sentence += fmt.Sprintf("It will rain at %s. ", strings.Join(willRain, ", "))
        }


        if len(meteoData.Data1H.PrecipitationProbability) > 0 {
            sentence += fmt.Sprintf("The current rain probability is %d percent.", meteoData.Data1H.PrecipitationProbability[currentTimeSpot])
        }

        willSnow := weatherdata.WillItSnow(meteoData)
        if len(willSnow) > 0 {
            sentence += fmt.Sprintf("It will snow at %s. ", strings.Join(willSnow, ", "))
        }

        willFog := weatherdata.WillItBeFoggy(meteoData)
        if len(willFog) > 0 {
            sentence += fmt.Sprintf("It will be foggy at %s. ", strings.Join(willFog, ", "))
        }

        willWind := weatherdata.WillItBeWindy(meteoData)
        if len(willWind) > 0 {
            sentence += fmt.Sprintf("It will be windy at %s. ", strings.Join(willWind, ", "))
        }

        highUVIndex := weatherdata.WillHaveHighUVIndex(meteoData)
        if len(highUVIndex) > 0 {
            sentence += fmt.Sprintf("There will be a high UV index at %s. ", strings.Join(highUVIndex, ", "))
        }

    }

    interpretedText := llm.GenerateSentence(sentence, body.Language)
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

    // err = unrealspeech.GenerateSpeech(models.SpeechRequest{
    //     Text:    interpretedText,
    //     VoiceId: "Scarlett",
    //     Bitrate: "64k",
    //     Speed:   "0",
    //     Pitch:   "1",
    //     Codec:   "libmp3lame",
    // }, filePath)

    err = tts.GoogleTextToSpeech(interpretedText, filePath, body.Language)

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

