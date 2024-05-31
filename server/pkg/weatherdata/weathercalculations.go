package weatherdata

import (
	"fmt"
	"server/pkg/models"
)



func AverageTemperature(data models.CityClimateData) float64 {
    var sum float64
    for _, feature := range data.Features {
        sum += feature.Properties.Values
    }
    return sum / float64(len(data.Features))
}


// temperatureNext1H calculates the average temperature for the next 3 hours given an array of float64 temperatures.
func TemperatureNext1H(data []float64) (float64, error) {
    if len(data) < 3 {
        return 0, fmt.Errorf("not enough data points to calculate the next 3 hours")
    }

    sum := 0.0
    for i := 0; i < 3; i++ {
        sum += data[i]
    }

    return sum / 3, nil
}


func PeakMeteoWindspeed(data models.MeteoBlueData) float64 {
    var max float64
    for _, windspeed := range data.Data1H.Windspeed {
        if windspeed > max {
            max = windspeed
        }
    }
    return max
}

func PeakMeteoTemperature(data models.MeteoBlueData) (float64, string) {
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
func WillItRain(data models.MeteoBlueData) ([]string) {
    var times []string
    for i, probability := range data.Data1H.PrecipitationProbability {
        if probability > 50 {
            times = append(times, data.Data1H.Time[i])
        }
    }
    return times
}

// willItSnow returns a slice of timestamps when the snow fraction is more than 0.5.
func WillItSnow(data models.MeteoBlueData) ([]string) {
    var times []string
    for i, snowFraction := range data.Data1H.SnowFraction {
        if snowFraction > 0.5 {
            times = append(times, data.Data1H.Time[i])
        }
    }
    return times
}

// willItBeFoggy returns a slice of timestamps when foggy conditions are detected (pictocode == 3).
func WillItBeFoggy(data models.MeteoBlueData) ([]string) {
    var times []string
    for i, pictocode := range data.Data1H.Pictocode {
        if pictocode == 3 {
            times = append(times, data.Data1H.Time[i])
        }
    }
    return times
}

// willItBeWindy returns a slice of timestamps when the windspeed exceeds 10.
func WillItBeWindy(data models.MeteoBlueData) ([]string) {
    var times []string
    for i, windspeed := range data.Data1H.Windspeed {
        if windspeed > 6 {
            times = append(times, data.Data1H.Time[i])
        }
    }
    return times
}


func WillHaveHighUVIndex(data models.MeteoBlueData) ([]string) {
    var times []string
    for i, uvIndex := range data.Data1H.UVIndex {
        if uvIndex > 4 {
            times = append(times, data.Data1H.Time[i])
        }
    }
    return times
}