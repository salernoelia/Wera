package models

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

type MeteoBlueData struct {
    Metadata Metadata `json:"metadata"`
    Data     Data     `json:"data_3h"` // Ensure this matches the JSON key for weather data
}

type Metadata struct {
    ModelRunUpdateTimeUTC string  `json:"modelrun_updatetime_utc"`
    Name                  string  `json:"name"`
    Height                int     `json:"height"`
    TimezoneAbbreviation  string  `json:"timezone_abbrevation"`
    Latitude              float64 `json:"latitude"`
    ModelRunUTC           string  `json:"modelrun_utc"`
    Longitude             float64 `json:"longitude"`
    UtcTimeOffset         float64 `json:"utc_timeoffset"` 
    GenerationTimeMs      float64 `json:"generation_time_ms"`
}


type Data struct {
    Time                   []string  `json:"time"`
    Temperature            []float64 `json:"temperature"`
    Windspeed              []float64 `json:"windspeed"`
    PrecipitationProbability []int   `json:"precipitation_probability"`
    // Include other necessary fields
}

type CityClimateData struct {
    Features []struct {
        Properties CityClimateSensor `json:"properties"`
    } `json:"features"`
}

type CityClimateSensor struct {
    Values float64 `json:"values"`
}

type TTSRequest struct {
    Text string `json:"text"`
}
