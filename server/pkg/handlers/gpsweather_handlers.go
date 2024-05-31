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

    var sentence string
    if closestSensor != nil {
        sentence = fmt.Sprintf("The closest sensor is %s, located %.2f km away. ", closestSensor.Name, distance)
    }

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

    if len(hotAreaNames) > 0 {
        sentence += fmt.Sprintf("Be aware of hot areas like %s. ", strings.Join(hotAreaNames, ", "))
    }

    sentence += fmt.Sprintf("The current time is %s and the date is %s. ", currentTime.Format("15:04"), currentTime.Format("2006-01-02"))
    sentence += "Generate a few sentences like a nice friend (dont claim to be one) packed around the data, sounding very personalized, without actually mentioning any numbers, except for the time and date"
    sentence += "If there is a high windspeed or UV Index, mention it and give hints to prevent sunburn or getting hurt in case of strong winds."
    sentence += "Only say things that fit to the actual weather no hypotheticals. So dont give useless advice like 'wear a jacket' if its 30 degrees outside. Or 'stay in the shade' if its raining."
    sentence += "Be friendly since you are talking to an elderly or non technical person, so do not mention any technicalities like sensors. If the temperatures are very high (over 30) please mention the risks of heat strokes and tell them to be cautious."
    sentence += "Please keep it short, maximum of 400 characters! Your name is Wera."

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


// temperatureNext1H calculates the average temperature for the next 3 hours given an array of float64 temperatures.
func temperatureNext1H(data []float64) (float64, error) {
    if len(data) < 3 {
        return 0, fmt.Errorf("not enough data points to calculate the next 3 hours")
    }

    sum := 0.0
    for i := 0; i < 3; i++ {
        sum += data[i]
    }

    return sum / 3, nil
}


func peakMeteoWindspeed(data models.MeteoBlueData) float64 {
    var max float64
    for _, windspeed := range data.Data1H.Windspeed {
        if windspeed > max {
            max = windspeed
        }
    }
    return max
}

func peakMeteoTemperature(data models.MeteoBlueData) (float64, string) {
    var max float64
    var timeOfMax string

    for i, temp := range data.Data1H.Temperature {
        if i == 0 || temp > max {  // Initialize max with the first element or update it
            max = temp
            timeOfMax = data.Data1H.Time[i]  // Assuming a corresponding Time slice
        }
    }

    return max, timeOfMax
}




// willItRain returns a slice of timestamps when the rain probability exceeds 50%.
func willItRain(data models.MeteoBlueData) ([]string) {
    var times []string
    for i, probability := range data.Data1H.PrecipitationProbability {
        if probability > 50 {
            times = append(times, data.Data1H.Time[i])
        }
    }
    return times
}

// willItSnow returns a slice of timestamps when the snow fraction is more than 0.5.
func willItSnow(data models.MeteoBlueData) ([]string) {
    var times []string
    for i, snowFraction := range data.Data1H.SnowFraction {
        if snowFraction > 0.5 {
            times = append(times, data.Data1H.Time[i])
        }
    }
    return times
}

// willItBeFoggy returns a slice of timestamps when foggy conditions are detected (pictocode == 3).
func willItBeFoggy(data models.MeteoBlueData) ([]string) {
    var times []string
    for i, pictocode := range data.Data1H.Pictocode {
        if pictocode == 3 {
            times = append(times, data.Data1H.Time[i])
        }
    }
    return times
}

// willItBeWindy returns a slice of timestamps when the windspeed exceeds 10.
func willItBeWindy(data models.MeteoBlueData) ([]string) {
    var times []string
    for i, windspeed := range data.Data1H.Windspeed {
        if windspeed > 6 {
            times = append(times, data.Data1H.Time[i])
        }
    }
    return times
}


func willHaveHighUVIndex(data models.MeteoBlueData) ([]string) {
    var times []string
    for i, uvIndex := range data.Data1H.UVIndex {
        if uvIndex > 4 {
            times = append(times, data.Data1H.Time[i])
        }
    }
    return times
}