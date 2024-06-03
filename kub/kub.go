package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
	"go.bug.st/serial.v1"
)

type GPSPayload struct {
    DeviceID  string  `json:"device_id"`
    Latitude  float64 `json:"Latitude"`
    Longitude float64 `json:"Longitude"`
    Language  string  `json:"Language"`
}
type Payload struct {
    DeviceID  string  `json:"device_id"`
    Language  string  `json:"Language"`
}


var lastValidLat, lastValidLon float64

var gpsActive bool = false




var GPSCheckLED rpio.Pin
var internetCheckLED rpio.Pin

 const (
    CLK = 2  // GPIO 2
	DT  = 3  // GPIO 3
    button   = 27  // GPIO 27 for the switch
    GPSLED = 26
    INTERNETLED =  13
    DEBOUNCE_DELAY = 100 * time.Millisecond // Debounce delay
    englishActive = "english-active.wav"
    germanActive = "german-active.wav"
    englishChangelang = "english-changelang.wav"
    germanChangelang = "german-changelang.wav"
    englishSelected = "english-selected.wav"
    germanSelected = "german-selected.wav"
)



var (
	inputDelta = 70
    languageInputDelta = 0
	printFlag  bool
	state      uint8
	encoderA   rpio.Pin
	encoderB   rpio.Pin
    lastVolumeChangeTime time.Time
    selectedLanguage string
)




func postAndPlayAudioGPS(url string, lat, lon float64, language string) {

    payload := GPSPayload{
        DeviceID:  "Lorena",
        Latitude:  lat,
        Longitude: lon,
        Language: language,
    }
    jsonData, err := json.Marshal(payload)
    if err != nil {
        fmt.Printf("Error marshaling JSON: %v\n", err)
        return
    }

    fmt.Println(bytes.NewBuffer(jsonData))

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Printf("Failed to create request: %v\n", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Printf("Failed to fetch audio: %v\n", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Printf("Failed to fetch audio: %d - %s\n", resp.StatusCode, resp.Status)
        return
    }

    audioData, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Failed to read audio data: %v\n", err)
        return
    }

    tmpFile, err := ioutil.TempFile("", "audio-*.wav")
	if err != nil {
		fmt.Printf("Failed to create a temp file: %v\n", err)
		return
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(audioData); err != nil {
		fmt.Printf("Failed to write to temp file: %v\n", err)
		return
	}

	fmt.Println("Playing audio...")
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ffmpeg -i %s -f wav - | aplay -D plughw:CARD=Headphones,DEV=0", tmpFile.Name()))
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to play audio: %v\n", err)
	}
}
func postAndPlayAudio(url string, language string) {

    payload := Payload{
        DeviceID:  "Lorena",
        Language: language,

    }

    jsonData, err := json.Marshal(payload)
    if err != nil {
        fmt.Printf("Error marshaling JSON: %v\n", err)
        return
    }

    fmt.Println(bytes.NewBuffer(jsonData))

    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Printf("Failed to create request: %v\n", err)
        return
    }
    req.Header.Set("Content-Type", "application/json")

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        fmt.Printf("Failed to fetch audio: %v\n", err)
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        fmt.Printf("Failed to fetch audio: %d - %s\n", resp.StatusCode, resp.Status)
        return
    }

    audioData, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Failed to read audio data: %v\n", err)
        return
    }

    tmpFile, err := ioutil.TempFile("", "audio-*.wav")
	if err != nil {
		fmt.Printf("Failed to create a temp file: %v\n", err)
		return
	}
	defer tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(audioData); err != nil {
		fmt.Printf("Failed to write to temp file: %v\n", err)
		return
	}

	fmt.Println("Playing audio...")
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ffmpeg -i %s -f wav - | aplay -D plughw:CARD=Headphones,DEV=0", tmpFile.Name()))
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to play audio: %v\n", err)
	}
}




func main() {
  
    if err := rpio.Open(); err != nil {
        fmt.Printf("Error opening GPIO: %v\n", err)
        return
    }
    defer rpio.Close()

    // Define the pins
 

    // Setup GPIO Pins for the encoder
    encoderA = rpio.Pin(CLK)
    encoderB = rpio.Pin(DT)

    encoderA.Input()
    encoderB.Input()
    encoderA.PullUp()
    encoderB.PullUp()

    // Setup the push button
    pushButton := rpio.Pin(button)
    pushButton.Input()
    pushButton.PullUp()

    // Initialize LEDs
    GPSCheckLED = rpio.Pin(GPSLED)
    GPSCheckLED.Output()
    GPSCheckLED.Low()

    internetCheckLED = rpio.Pin(INTERNETLED)
    internetCheckLED.Output()
    internetCheckLED.Low()

    setupCloseHandler()



    mode := &serial.Mode{
    BaudRate: 9600,
    Parity:   serial.NoParity,
    DataBits: 8,
    StopBits: serial.OneStopBit,
}

port, err := serial.Open("/dev/ttyAMA0", mode)
if err != nil {
    fmt.Println("Failed to open serial port:", err)
    return
}
defer port.Close()

reader := bufio.NewReader(port)
var gpsDataBlock string




  go func() {
        for {
            c, err := reader.ReadByte()
            if err != nil {
                if err == syscall.EINTR {
                    continue // Handle interrupted system call
                }
                fmt.Println("Failed to read from serial port:", err)
                time.Sleep(100 * time.Millisecond)
                continue
            }

            // Convert the byte to string
            char := string(c)

            if char == "\n" {

                if strings.HasPrefix(gpsDataBlock, "$GPRMC") {
                    processGPSData(gpsDataBlock)
                }
                gpsDataBlock = ""
            } else {
                gpsDataBlock += char
            }
        }
    }()

    // Set the language to German or English
    setLanguage(&pushButton, &encoderA, &encoderB)



    // Asynchronous Routine to check internet connection
    go checkInternetConnectivityPeriodically()

    // Button press handling in a separate goroutine
    go handleButtonPress(&pushButton)

    // Set initial volume to a midpoint (e.g., 50%)
	setVolume(80)

  
    for {
        readEncoder()
        printDelta()
        time.Sleep(1 * time.Millisecond) // Small sleep to prevent high CPU load
    }

}

func playAudio(inputFile string) {

    cmd := exec.Command("aplay", "-D", "plughw:CARD=Headphones,DEV=0", inputFile)
    var stdout, stderr bytes.Buffer
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr

    if err := cmd.Run(); err != nil {
        log.Printf("Failed to play audio: %s\nStderr: %s\n", err, stderr.String())
    } else {
        fmt.Printf("Audio played successfully. Output: %s\n", stdout.String())
    }
}



func setLanguage(button *rpio.Pin, encoderA *rpio.Pin, encoderB *rpio.Pin) {
    // var currentLanguage string = "english" // default language
	setVolume(90)

	fmt.Println(100)
    selectedLanguage = "german"
    // languageCode = 0


    fmt.Println("Playing german init...")
    playAudio(germanChangelang)
  

    fmt.Println("Playing english init...")
    playAudio(englishChangelang)

    for {

        if encoderA.Read() == rpio.Low {
            fmt.Println(languageInputDelta)

                fmt.Printf("Selected Language: %s\n", "Deutsch")
                selectedLanguage = "german"
                playAudio(germanActive)
        } else if encoderB.Read() == rpio.Low {
                fmt.Printf("Selected Language: %s\n", "English")
                selectedLanguage = "english"
                playAudio(englishActive)
        }

        time.Sleep(1 * time.Millisecond) // Small sleep to prevent high CPU load


        // Check if the button is pressed to confirm the selection
        if button.Read() == rpio.Low {

            if selectedLanguage == "german" {
                playAudio(germanSelected)
                fmt.Print("Button is pressed")
                time.Sleep(10 * time.Millisecond) // Polling delay to reduce CPU usag
                return
            } else if selectedLanguage == "english" {
                playAudio(englishSelected)
                fmt.Print("Button is pressed")
                time.Sleep(10 * time.Millisecond) // Polling delay to reduce CPU usag
                return
            }
           

        }

        // time.Sleep(DEBOUNCE_DELAY) // Sleep to reduce CPU usage and debounce handling
    }
}


// Function to handle button presses asynchronously with improved debouncing
func handleButtonPress(button *rpio.Pin) {

    for {
        if button.Read() == rpio.Low {
			fmt.Println("Button pressed!")
             handleButtonActions()
			time.Sleep(time.Second) // Add a delay to debounce the button press
		}
		time.Sleep(10 * time.Millisecond) // Polling delay to reduce CPU usag

    }
}


// Function to handle actions upon button press
func handleButtonActions() {
    isConnected := checkInternetConnection("https://spatial-interaction.onrender.com/ok")
    if gpsActive && isConnected {
        fmt.Println("Trying to post to /weathergps...", lastValidLat, lastValidLon)
        postAndPlayAudioGPS("https://spatial-interaction.onrender.com/weathergps", lastValidLat, lastValidLon, selectedLanguage)
    } else if !gpsActive && isConnected {
        fmt.Println("Trying to get to /weather since no GPS data is available...")
        postAndPlayAudio("https://spatial-interaction.onrender.com/weather", selectedLanguage)
    } else {
        fmt.Println("Internet Status:", isConnected)
        fmt.Println("GPS Status:", gpsActive)
    }
}

// Periodically checks internet connection
func checkInternetConnectivityPeriodically() {
    urlToCheck := "https://spatial-interaction.onrender.com/ok"
    fmt.Println("Checked OK")
    for {
        isConnected := checkInternetConnection(urlToCheck)
        if isConnected {
            internetCheckLED.Low()
        } else {
            internetCheckLED.High()
        }
        time.Sleep(200 * time.Second) // Check every 200 seconds
    }
}


func checkInternetConnection(url string) bool {
    timeout := time.Duration(80 * time.Second)
    client := http.Client{
        Timeout: timeout,
    }
    _, err := client.Get(url)
    if err != nil {
        fmt.Println("Error checking internet connection:", err)
        internetCheckLED.High()
        return false
    }
    return true
}

func readEncoder() {
	CLKstate := encoderA.Read() == rpio.Low
	DTstate := encoderB.Read() == rpio.Low

	switch state {
	case 0: // Idle state, encoder not turning
		if !CLKstate { // Turn clockwise and CLK goes low first
			state = 1
		} else if !DTstate { // Turn anticlockwise and DT goes low first
			state = 4
		}
	case 1:
		if !DTstate { // Continue clockwise and DT will go low after CLK
			state = 2
		}
	case 2:
		if CLKstate { // Turn further and CLK will go high first
			state = 3
		}
	case 3:
		if CLKstate && DTstate { // Both CLK and DT now high as the encoder completes one step clockwise
			state = 0
			inputDelta++
			printFlag = true
		}
	case 4:
		if !CLKstate {
			state = 5
		}
	case 5:
		if DTstate {
			state = 6
		}
	case 6:
		if CLKstate && DTstate {
			state = 0
			inputDelta--
			printFlag = true
		}
	}
}




func printDelta() {
	if printFlag && time.Since(lastVolumeChangeTime) > DEBOUNCE_DELAY {
		printFlag = false
		volume := mapDeltaToVolume(int(inputDelta))
		fmt.Println(volume)
		setVolume(volume)
		lastVolumeChangeTime = time.Now()
	}
}



func mapDeltaToVolume(delta int) int {
	// Map inputDelta to a range of 0 to 100
	if delta < 0 {
		delta = 0
	} else if delta > 100 {
		delta = 100
	}
	return delta
}


func setVolume(volume int) error {
	cmd := exec.Command("amixer", "cset", "numid=1", fmt.Sprintf("%d%%", volume))
	return cmd.Run()
}


func processGPSData(data string) {
    if strings.HasPrefix(data, "$GPRMC") {
        fields := strings.Split(data, ",")
        if len(fields) > 6 && fields[2] == "A" { // Check for 'Active' status
            latitude := fields[3]
            ns := fields[4]
            longitude := fields[5]
            ew := fields[6]

            lat := convertToDecimalDegrees(latitude, true)
            lon := convertToDecimalDegrees(longitude, false)

            if ns == "S" {
                lat = -lat
            }
            if ew == "W" {
                lon = -lon
            }

            if lat != 0.0 && lon != 0.0 {
                lastValidLat = lat
                lastValidLon = lon
                gpsActive = true
                GPSCheckLED.High()
                fmt.Printf("Valid GPS Data: Latitude: %.6f, Longitude: %.6f\n", lat, lon)
            } else {
                fmt.Println("Invalid GPS Data: Skipping zero coordinates.")
            }
        } else {
            fmt.Println("Invalid GPS Data: No 'Active' status.")
        }
    }
}


func convertToDecimalDegrees(coordinate string, isLatitude bool) float64 {
    degreeLength := 2
    if !isLatitude {
        degreeLength = 3
    }
    
    // Check if the coordinate has enough characters to avoid out-of-range error
    if len(coordinate) < degreeLength {
        fmt.Printf("Invalid coordinate data: %s\n", coordinate)
        return 0
    }
    
    var degrees, minutes float64
    fmt.Sscanf(coordinate[:degreeLength], "%f", &degrees)
    fmt.Sscanf(coordinate[degreeLength:], "%f", &minutes)
    return degrees + minutes/60
}

func setupCloseHandler() {
    c := make(chan os.Signal)
    signal.Notify(c, os.Interrupt, syscall.SIGTERM)
    go func() {
        <-c
        fmt.Println("\nCTRL+C pressed. Cleaning up and exiting...")
        cleanup()
        os.Exit(0)
    }()
}

func cleanup() {
    fmt.Println("Turning off...")
    GPSCheckLED.Low() // Turn off LED
    // Add any other cleanup tasks here
}