package handlers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"server/pkg/weatherdata"
)

func TestHandler(w http.ResponseWriter, r *http.Request) {
    // Fetch MeteoBlue data
    meteoData, err := weatherdata.FetchMeteoBlueData()
    if err != nil {
        log.Printf("Error fetching MeteoBlue data: %v\n", err)
        http.Error(w, "Failed to fetch MeteoBlue data", http.StatusInternalServerError)
        return
    }

    // Switzerland is typically in CET or CEST, so let's assume CEST for now (+2 UTC)
    location, err := time.LoadLocation("Europe/Zurich")
    if err != nil {
        log.Printf("Error loading location 'Europe/Zurich': %v", err)
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }
    currentTime := time.Now().In(location)
    sentence := fmt.Sprintf("The current local time is %s and the date is %s. ", currentTime.Format("15:04"), currentTime.Format("2006-01-02"))
    fmt.Fprint(w, sentence)

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
}
