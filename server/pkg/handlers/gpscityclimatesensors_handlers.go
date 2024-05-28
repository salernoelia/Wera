package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"server/pkg/handlegps"
	"server/pkg/models"
	"server/pkg/weatherdata"
)


func ListCityClimateSensorsBasedOnDistance(w http.ResponseWriter, r *http.Request) {
    var body models.RadioRequestBody
    err := json.NewDecoder(r.Body).Decode(&body)
    if err != nil {
        http.Error(w, "Invalid JSON input", http.StatusBadRequest)
        return
    }

    cityClimateData, err := weatherdata.FetchCityClimateData()
    if err != nil {
        log.Printf("Error fetching CityClimate data: %v\n", err)
        http.Error(w, "Failed to fetch CityClimate data", http.StatusInternalServerError)
        return
    }

    // list of sensors closest to furthest
	sensors := handlegps.FindClosestSensorList(cityClimateData, body.Latitude, body.Longitude)

	 w.Header().Set("Content-Type", "application/json")
     if err := json.NewEncoder(w).Encode(sensors); err != nil {
        log.Printf("Error encoding response: %v", err)
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
        return
    }
	
}

