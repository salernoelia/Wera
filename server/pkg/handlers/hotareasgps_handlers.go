package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/pkg/handlegps"
	"server/pkg/models"
	"server/pkg/weatherdata"
)
func FetchAndReportHotAreasBasedOnLocation(w http.ResponseWriter, r *http.Request) {
    var body models.RadioRequestBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        http.Error(w, "Invalid JSON input: "+err.Error(), http.StatusBadRequest)
        return
    }

    cityClimateData, err := weatherdata.FetchCityClimateData()
    if err != nil {
        log.Printf("Error fetching CityClimate data: %v", err)
        http.Error(w, "Failed to fetch CityClimate data", http.StatusInternalServerError)
        return
    }

    hotAreas, err := weatherdata.FindHotAreas(cityClimateData)
    if err != nil {
        http.Error(w, "No hot areas found: "+err.Error(), http.StatusNotFound)
        return
    }

    hotAreasBasedOnLocation := handlegps.FindClosestSensorSlicedList(hotAreas, body.Latitude, body.Longitude)

    // Location names of hot areas
    var hotAreaNames []string
    for _, area := range hotAreasBasedOnLocation {
        hotAreaNames = append(hotAreaNames, area.Name)
    }

    log.Println(hotAreaNames)

    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(hotAreasBasedOnLocation); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}
