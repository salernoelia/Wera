package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"server/pkg/weatherdata"
)



func FetchAndReportHotAreas(w http.ResponseWriter, r *http.Request) {
	cityClimateData, err := weatherdata.FetchCityClimateData()
	if err != nil {
		log.Printf("Error fetching CityClimate data: %v\n", err)
		http.Error(w, "Failed to fetch CityClimate data", http.StatusInternalServerError)
		return
	}

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

	fmt.Println(hotAreaNames)


	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hotAreas)
}
