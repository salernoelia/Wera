package handlegps

import (
	"math"
	"server/pkg/models"
	"sort"
)

// Calculate the closest sensor based on Euclidean distance
func FindClosestSensor(data models.CityClimateData, lat, lon float64) (*models.CityClimateSensor, float64) {
    var closest *models.CityClimateSensor
    minDistance := math.MaxFloat64
    for _, feature := range data.Features {
        sensorLat := feature.Geometry.Coordinates.Lat
        sensorLon := feature.Geometry.Coordinates.Lon
        distance := HaversineDistance(lat, lon, sensorLat, sensorLon)
        if distance < minDistance {
            minDistance = distance
            closest = &feature.Properties
        }
    }
    return closest, minDistance
}


// Calculate Haversine distance between two points
func HaversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
    var R float64 = 6371 // Radius of the Earth in km
    var dLat = (lat2 - lat1) * math.Pi / 180
    var dLon = (lon2 - lon1) * math.Pi / 180
    var a = math.Sin(dLat/2)*math.Sin(dLat/2) +
        math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
        math.Sin(dLon/2)*math.Sin(dLon/2)
    var c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
    return R * c // Distance in km
}

// FindClosestSensorList returns a sorted list of sensors from closest to furthest
func FindClosestSensorList(data models.CityClimateData, lat, lon float64) []models.CityClimateSensorDistance {
    sensors := make([]models.CityClimateSensorDistance, len(data.Features))

    for i, feature := range data.Features {
        dist := HaversineDistance(lat, lon, feature.Geometry.Coordinates.Lat, feature.Geometry.Coordinates.Lon)

        sensors[i] = models.CityClimateSensorDistance{
            CityClimateSensor: feature.Properties,
            Distance:          dist,
            Geometry:          feature.Geometry, // Ensure the geometry is copied
        }
    }

    // Sort the sensors based on the computed distance
    sort.Slice(sensors, func(i, j int) bool {
        return sensors[i].Distance < sensors[j].Distance
    })

    return sensors
}



// FindClosestSensorSlicedList calculates the distance for a slice of sensors and returns them sorted by distance.
func FindClosestSensorSlicedList(sensors []models.CityClimateSensor, lat, lon float64) []models.CityClimateSensorDistance {
    sensorDistances := make([]models.CityClimateSensorDistance, len(sensors))

    for i, sensor := range sensors {
        distance := HaversineDistance(lat, lon, sensor.Geometry.Coordinates.Lat, sensor.Geometry.Coordinates.Lon)
        
        // Now we include the geometry in the distance struct
        sensorDistances[i] = models.CityClimateSensorDistance{
    CityClimateSensor: sensor,
    Distance: distance,
    Geometry: sensor.Geometry,  // Ensure geometry is correctly referenced here
}
    }

    sort.Slice(sensorDistances, func(i, j int) bool {
        return sensorDistances[i].Distance < sensorDistances[j].Distance
    })

    return sensorDistances
}



