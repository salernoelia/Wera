package weatherdata

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"server/pkg/models"
	"strconv"
	"time"
)

func FetchMeteoBlueData() (models.MeteoBlueData, error) {
	meteoApi := os.Getenv("METEO_API_KEY")
	if meteoApi == "" {
		log.Fatal("API_KEY environment variable is not set.")
	}
	url := "https://my.meteoblue.com/packages/basic-3h?apikey=" + meteoApi + "&lat=47.3667&lon=8.55&asl=429&format=json"

	resp, err := http.Get(url)
	if err != nil {
		return models.MeteoBlueData{}, err
	}
	defer resp.Body.Close()

	var meteoData models.MeteoBlueData
	if err := json.NewDecoder(resp.Body).Decode(&meteoData); err != nil {
		return models.MeteoBlueData{}, err
	}

	if len(meteoData.Data.Temperature) == 0 {
		return models.MeteoBlueData{}, errors.New("no temperature data received from MeteoBlue API")
	}

	return meteoData, nil
}

func FetchCityClimateData() (models.CityClimateData, error) {
	now := time.Now()
	unixTimestamp := now.Unix()
	unixTimestampRoundedToHour := (unixTimestamp / 3600) * 3600

	meteoApi := os.Getenv("METEO_API_KEY")
	if meteoApi == "" {
		log.Fatal("API_KEY environment variable is not set.")
	}
	url := "https://www.meteoblue.com/de/products/cityclimate/getData?locationId=2657896&type=temperature&units=m&time=" + strconv.FormatInt(unixTimestampRoundedToHour, 10) + "&apikey=" + meteoApi

	resp, err := http.Get(url)
	if err != nil {
		return models.CityClimateData{}, err
	}
	defer resp.Body.Close()

	 var cityClimateData models.CityClimateData
    if err := json.NewDecoder(resp.Body).Decode(&cityClimateData); err != nil {
        return models.CityClimateData{}, err
    }
    
    // Debugging to check data integrity
    // log.Printf("Received city climate data: %+v", cityClimateData)
    
    return cityClimateData, nil
}

// FindHotAreas returns a JSON of all locations with temperatures exceeding 28 degrees Celsius.
func FindHotAreas(data models.CityClimateData) ([]models.CityClimateSensor, error) {
    var hotAreas []models.CityClimateSensor
    for _, feature := range data.Features {
        if feature.Properties.Values > 28 { // assuming 28°C is the threshold for hot areas
            completeSensor := feature.Properties
            completeSensor.Geometry = feature.Geometry // Make sure geometry is included
            hotAreas = append(hotAreas, completeSensor)
        }
    }

    if len(hotAreas) == 0 {
        return nil, fmt.Errorf("no hot areas found")
    }

    return hotAreas, nil
}