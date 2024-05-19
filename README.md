# Spatial Interaction Radio

Backend for the spatial interaction module project, name is still undefined.

Be sure to have a Postgres DB setup.

### Install eSpeak

```bash
brew install espeak
```

### Start the server for testing:

```bash
go run cmd/risque-server/main.go
```

### Compile to exec:

```bash
go build cmd/risque-server/main.go
```

---

## Raspberry PI Setup

#### To build get the dependency and build

```bash
go get github.com/stianeikeland/go-rpio/v4

go build RaspberryPI/v2/goradio.go
```

---

# Docs

## Endpoints

### **GET /cityclimate**

Responds with the sensor dataset of the ZHAW Grid, currently only has access to about 50 sensors and the temperature data only.

### **GET /meteoblue**

Responds with a 3-hour forecast from Meteoblue data, also provides a 24-hour overview. Data contains temperature, wind, rain, and some more. Each request takes 8000 tokens and our free API is limited to 10M, so please make only as many requests as needed.

### **POST /users**

Create a user with the following format:

```JSON
{
 "name": "John Shoe",
 "email": "john@example.com"
}
```

### **GET /users**

Responds with JSON of all users.

### **GET /weather**

Fetches and combines data from both the MeteoBlue and CityClimate APIs, processes it to compute an average temperature, constructs a descriptive sentence, and generates a speech file which is returned as an audio stream.

The response contains:

- The current average temperature of the sensor grid.
- The temperature and wind speed according to MeteoBlue.

### Example Response:

```plaintext
The current average temperature of the Sensor Grid is 22.50 degrees Celsius. According to MeteoBlue, the temperature is 20.10 degrees Celsius with a windspeed of 3.5 meters per second.
```

### Example Request in Python:

```python
import requests
from pydub import AudioSegment
from pydub.playback import play
import io

def fetch_and_play_audio(url):
    response = requests.get(url)
    if response.status_code == 200:
        audio_data = io.BytesIO(response.content)
        song = AudioSegment.from_file(audio_data, format="wav")
        print("Playing audio...")
        play(song)
    else:
        print("Failed to fetch audio:", response.status_code)

fetch_and_play_audio("http://192.168.1.13:8080/weather")
```

### Dependencies

- `github.com/jackc/pgx/v4` for PostgreSQL database interaction.
- `github.com/gorilla/mux` for routing.
- External TTS library for text-to-speech conversion.
