# Spatial Interaction Radio

Backend for the spatial interaction module project, name is still undefined.

### Start the server for testing:

```bash
go run cmd/server/main.go
```

### Compile to exec:

```bash
go build cmd/server/main.go
```

---

## Raspberry PI Setup

### Fetch dependencies and build:

```bash
go get github.com/stianeikeland/go-rpio/v4

go build RaspberryPI/v2/goradio.go
```

---

# Docs

[Selfhosted Service](http://estationserve.ddns.net:9000/)
OR
[Render Hosted Service (1min startup time)](https://spatial-interaction.onrender.com)

## Endpoints

### **GET /cityclimate**

Responds with the sensor dataset of the ZHAW Grid, currently only has access to about 50 sensors and the temperature data only.

## **POST /cityclimategps**

Responds with the sensor dataset of the ZHAW Grid, sorted by distance (closest to furthest).

```JSON
// Sample Request Body
{
  "device_id": "Device_1",
  "Latitude": 47.3653466,
  "Longitude": 8.5282651
}
```

```JSON
// Response
[
  {
    "id": "03400120",
    "name": "Sihlhölzlistrasse",
    "timestamp": "1716555600",
    "values": 22.8375,
    "colors": "#9baf33",
    "active": 1,
    "Distance": 0.3995830946586305,
    "geometry": {
      "type": "point",
      "coordinates": {
        "lon": 8.528826,
        "lat": 47.36892
      }
    }
  },
...
]
```

### **GET /meteoblue**

Responds with a 3-hour forecast from Meteoblue data, also provides a 24-hour overview. Data contains temperature, wind, rain, and some more. Each request takes 8000 tokens and our free API is limited to 10M, so please make only as many requests as needed.

## **POST /weathergps**

Fetches and combines data from both the MeteoBlue and CityClimate APIs, processes it to compute an average temperature, processes GPS data to deliver data for the current location of the radio, constructs a nice response, and generates a speech file (.MP3) which is returned as an audio stream.

```JSON
// Sample Request Body
{
  "device_id": "Device_1",
  "Latitude": 47.3653466,
  "Longitude": 8.5282651
}
```

```plaintext
// samle text it generates and speaks out
"Good morning! It's a lovely day outside! The temperature has been quite pleasant, a gentle breeze is blowing and the temperature is just right - not too hot, not too cold. Just perfect. Make sure to stay hydrated and take a break if you're spending time outdoors. And remember, on especially warm days, please be extra careful to avoid heat exhaustion. Stay cool and comfortable!"
```


### **GET /weather**

Fetches and combines data from both the MeteoBlue and CityClimate APIs, processes it to compute an average temperature, constructs a descriptive sentence, and generates a speech file (.MP3) which is returned as an audio stream.

The response contains:

- The current average temperature of the sensor grid.
- The temperature and wind speed according to MeteoBlue.

```plaintext
// Speech file text
The current average temperature of the Sensor Grid is 22.50 degrees Celsius. According to MeteoBlue, the temperature is 20.10 degrees Celsius with a windspeed of 3.5 meters per second.
```

### **GET /hotareas**

Resonds with an array of sensors that are over a certain set threshold, on default it is set to 28 degrees celsius.

```json
// Response
[
  {
    "id": "0340011C",
    "name": "Zollikerstrasse",
    "timestamp": "1716555600",
    "values": 28.8025,
    "colors": "#743933",
    "active": 1,
    "geometry": {
      "type": "point",
      "coordinates": {
        "lon": 8.556188,
        "lat": 47.3617
      }
    }
  }
]
```

### **POST /hotareasgps**

Resonds with an array of sensors that are over a certain set threshold, on default it is set to 28 degrees celsius, sorted by distance to the post location.


```JSON
// Sample Request Body
{
  "device_id": "Device_1",
  "Latitude": 47.3653466,
  "Longitude": 8.5282651
}
```

```json
// Response
[
  {
    "id": "0340011C",
    "name": "Zollikerstrasse",
    "timestamp": "1716555600",
    "values": 28.8025,
    "colors": "#743933",
    "active": 1,
    "geometry": {
      "type": "point",
      "coordinates": {
        "lon": 8.556188,
        "lat": 47.3617
      }
    }
  }
]
```

### **POST /speak**

Text to speech endpoint. (Unreal speech)

```JSON
// Request body
{
    "Text": "Hello, this is a test of the Unreal Speech API integration. How does this sound?",
    "VoiceId": "Amy",
    "Bitrate": "64k",
    "Speed": "0",
    "Pitch": "1",
    "Codec": "libmp3lame"
}
```

```JSON
// Response Body
{
  "file": "output.mp3",
  "message": "Speech generated successfully"
}
```




### Dependencies

- `github.com/jackc/pgx/v4` for PostgreSQL database interaction.
- `github.com/gorilla/mux` for routing.
- External TTS library for text-to-speech conversion.
