# Spatial Interaction Radio

Software for the spatial interaction module project, name is still undefined.

## Table of Contents

- [Server Quickstart](#server-quickstart)
- [Radio (Raspberry PI 4) Quickstart](#radio-raspberry-pi-4-quickstart)
  - [Fetch dependencies and build](#fetch-dependencies-and-build)
- [Docs](#docs)
  - [Selfhosted Service](#selfhosted-service)
  - [Endpoints](#endpoints)
    - [GET /cityclimate](#get-cityclimate)
    - [POST /cityclimategps](#post-cityclimategps)
    - [GET /meteoblue](#get-meteoblue)
    - [POST /weathergps](#post-weathergps)
    - [GET /weather](#get-weather)
    - [GET /hotareas](#get-hotareas)
    - [POST /hotareasgps](#post-hotareasgps)
    - [POST /speak](#post-speak)
- [Dependencies](#dependencies)
  - [Server](#server-1)
  - [Radio Device (Raspberry Pi 4)](#radio-device-raspberry-pi-4)

## Server Quickstart:

Be sure to get all API Keys, .env is formatted like this

```
METEO_API_KEY=key
TTS_API_KEY=key
GROQ_API_KEY=key
UNREAL_SPEECH_API_KEY=key
```

1. **Navigate to the server directory:**
    ```bash
    cd server
    ```
2. **Initialize the Go module** (if not already done):
    ```bash
    go mod init server
    ```
3. **Fetch dependencies:**
    ```bash
    go get github.com/jackc/pgx/v4 github.com/gorilla/mux
    ```
4. **Build the server:**
    ```bash
    go build cmd/server/main.go
    ```

## Radio (Raspberry PI 4) Quickstart

### Fetch dependencies and build:

1. **Navigate to the radio directory:**
    ```bash
    cd radio
    ```
2. **Initialize the Go module** (if not already done):
    ```bash
    go mod init goradio
    ```
3. **Fetch dependencies:**
    ```bash
    go get github.com/stianeikeland/go-rpio/v4 go.bug.st/serial.v1
    ```
4. **Build the radio application:**
    ```bash
    go build goradio.go
    ```
5. **Run the built application:**
    ```bash
    ./goradio
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

Sample Request Body:
```JSON
{
  "device_id": "Device_1",
  "Latitude": 47.3653466,
  "Longitude": 8.5282651
}
```

Sample Response Body:
```JSON
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
  }
]
```


### **GET /meteoblue**

Responds with a 3-hour forecast from Meteoblue data, also provides a 24-hour overview. Data contains temperature, wind, rain, and some more. Each request takes 8000 tokens and our free API is limited to 10M, so please make only as many requests as needed.

### **POST /weathergps**

Fetches and combines data from both the MeteoBlue and CityClimate APIs, processes it to compute an average temperature, processes GPS data to deliver data for the current location of the radio, constructs a nice response, and generates a speech file (.MP3) which is returned as an audio stream.

Sample Request Body:
```JSON
{
  "device_id": "Device_1",
  "Latitude": 47.3653466,
  "Longitude": 8.5282651
}
```

Sample text it generates and speaks:
```plaintext
"Good morning! It's a lovely day outside! The temperature has been quite pleasant, a gentle breeze is blowing and the temperature is just right - not too hot, not too cold. Just perfect. Make sure to stay hydrated and take a break if you're spending time outdoors. And remember, on especially warm days, please be extra careful to avoid heat exhaustion. Stay cool and comfortable!"
```


### **GET /weather**

Fetches and combines data from both the MeteoBlue and CityClimate APIs, processes it to compute an average temperature, constructs a descriptive sentence, and generates a speech file (.MP3) which is returned as an audio stream. This endpoint is used as a relay in case weathergps fails.

The response contains:

- The current average temperature of the sensor grid.
- The temperature and wind speed according to MeteoBlue.

Sample text it generates and speaks:
```plaintext
The current average temperature of the Sensor Grid is 22.50 degrees Celsius. According to MeteoBlue, the temperature is 20.10 degrees Celsius with a windspeed of 3.5 meters per second.
```


### **GET /hotareas**


Resonds with an array of sensors that are over a certain set threshold, on default it is set to 28 degrees celsius.

Sample Response Body:
```json
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



### **POST /hotareasgps**


Resonds with an array of sensors that are over a certain set threshold, on default it is set to 28 degrees celsius, sorted by distance to the post location.

Sample Request Body:
```JSON
{
  "device_id": "Device_1",
  "Latitude": 47.3653466,
  "Longitude": 8.5282651
}
```

Sample Response Body:
```json
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


### **POST /speak**

Text to speech endpoint. (Unreal speech)

Sample Request Body:
```JSON
{
    "Text": "Hello, this is a test of the Unreal Speech API integration. How does this sound?",
    "VoiceId": "Amy",
    "Bitrate": "64k",
    "Speed": "0",
    "Pitch": "1",
    "Codec": "libmp3lame"
}
```

Sample Response Body:
```JSON
{
  "file": "output.mp3",
  "message": "Speech generated successfully"
}
```

---

## Dependencies

### Server
- `github.com/jackc/pgx/v4` for PostgreSQL database interaction.
- `github.com/gorilla/mux` for routing.


### Radio Device (Raspberry Pi 4)
- `github.com/stianeikeland/go-rpio/v4`for GPIO pin support.
- `go.bug.st/serial.v1`for Serial Support.


### API's
- [MeteoBlue](https://www.meteoblue.com/de/weather-api/index/overview)
- [VoiceRSS API (backup TTS)](https://www.voicerss.org/personel/)
- [Groq](https://console.groq.com/docs/quickstart)
- [Unreal TTS](https://unrealspeech.com/onboard)
