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

// CityClimateData represents the main structure for the climate data API response.
type CityClimateData struct {
    Type   string `json:"type"`
    Series string `json:"series"`
    Scale  struct {
        ValuesMin float64 `json:"values_min"`
        ValuesMax float64 `json:"values_max"`
    } `json:"scale"`
    Meta struct {
        TimezoneOffsetS     int    `json:"timezone_offset_s"`
        TimezoneAbbreviation string `json:"timezone_abbreviation"`
        LocalFirst          int64  `json:"local_first"`
        LocalLast           int64  `json:"local_last"`
        Unit                string `json:"unit"`
    } `json:"meta"`
    Features []struct {
        Type     string `json:"type"`
        Geometry struct {
            Type        string `json:"type"`
            Coordinates struct {
                Lon float64 `json:"lon"`
                Lat float64 `json:"lat"`
            } `json:"coordinates"`
        } `json:"geometry"`
        Properties CityClimateSensor `json:"properties"`
    } `json:"features"`
}

// CityClimateSensor represents the sensor data for a specific location.
type CityClimateSensor struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    Timestamp string  `json:"timestamp"`
    Values    float64 `json:"values"`
    Colors    string  `json:"colors"`
    Active    int     `json:"active"`
    Geometry  struct {
        Type        string    `json:"type"`
        Coordinates struct {
            Lon float64 `json:"lon"`
            Lat float64 `json:"lat"`
        } `json:"coordinates"`
    } `json:"geometry"`
}





// CityClimateSensorDistance extends CityClimateSensor with a Distance field
type CityClimateSensorDistance struct {
    CityClimateSensor
    Distance float64
    Geometry struct {
        Type        string    `json:"type"`
        Coordinates struct {
            Lon float64 `json:"lon"`
            Lat float64 `json:"lat"`
        } `json:"coordinates"`
    } `json:"geometry"`
}



type TTSRequest struct {
    Text string `json:"text"`
}

type APIResponse struct {
    Choices []struct {
        Message struct {
            Role    string `json:"role"`
            Content string `json:"content"`
        } `json:"message"`
    } `json:"choices"`
}

type SpeechRequest struct {
    Text    string `json:"Text"`
    VoiceId string `json:"VoiceId"`
    Bitrate string `json:"Bitrate"`
    Speed   string `json:"Speed"`
    Pitch   string `json:"Pitch"`
    Codec   string `json:"Codec"`
}

type RadioRequestBody struct {
    DeviceID  string  `json:"device_id"`
    Latitude  float64 `json:"latitude"`
    Longitude float64 `json:"longitude"`
}