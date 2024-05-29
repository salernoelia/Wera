package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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
}



var lastValidLat, lastValidLon float64

var gpsActive bool = false

var GPSCheckLED rpio.Pin
var internetCheckLED rpio.Pin

 const (
    CLK = 2  // GPIO 2
	DT  = 3  // GPIO 3
    button   = 22  // GPIO 22 for the switch
    GPSLED = 26
    INTERNETLED =  13
    DEBOUNCE_DELAY = 100 * time.Millisecond // Debounce delay
)



var (
	inputDelta = 70
	printFlag  bool
	state      uint8
	encoderA   rpio.Pin
	encoderB   rpio.Pin
    lastVolumeChangeTime time.Time
    
)


func postAndPlayAudio(url string, lat, lon float64) {
    payload := GPSPayload{
        DeviceID:  "Device_1",
        Latitude:  lat,
        Longitude: lon,
    }
    jsonData, err := json.Marshal(payload)
    if err != nil {
        fmt.Printf("Error marshaling JSON: %v\n", err)
        return
    }

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
    cmd := exec.Command("ffplay", "-nodisp", "-autoexit", tmpFile.Name())
    if err := cmd.Run(); err != nil {
        fmt.Printf("Failed to play audio: %v\n", err)
    }
}

func getAndPlayAudio(url string) {
    resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Failed to fetch audio: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch audio: %d\n", resp.StatusCode)
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

            if c == '\n' {
                if strings.HasPrefix(gpsDataBlock, "$GNRMC") {
                    processGPSData(gpsDataBlock)
                }
                gpsDataBlock = ""
            } else {
                gpsDataBlock += string(c)
            }
        }
    }()

    // Asynchronous Routine to check internet connection
    go checkInternetConnectivityPeriodically()

    // Button press handling in a separate goroutine
    go handleButtonPress(&pushButton)

    // Set initial volume to a midpoint (e.g., 50%)
	setVolume(50)

  
    for {
        readEncoder()
        printDelta()
        time.Sleep(1 * time.Millisecond) // Small sleep to prevent high CPU load
    }

}

// Function to handle button presses asynchronously
func handleButtonPress(button *rpio.Pin) {
    for {
        if button.Read() == rpio.Low {
            fmt.Println("Button Pressed")
            handleButtonActions()
            time.Sleep(1000 * time.Millisecond) // Debounce delay
        }
        time.Sleep(10 * time.Millisecond) // Check button state at a regular interval
    }
}

// Function to handle actions upon button press
func handleButtonActions() {
    isConnected := checkInternetConnection("http://www.google.com")
    if gpsActive && isConnected {
        fmt.Println("Trying to post to /weathergps...")
        postAndPlayAudio("https://spatial-interaction.onrender.com/weathergps", lastValidLat, lastValidLon)
    } else if !gpsActive && isConnected {
        fmt.Println("Trying to get to /weather since no GPS data is available...")
        getAndPlayAudio("https://spatial-interaction.onrender.com/weather")
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
        time.Sleep(200 * time.Second) // Check every 10 seconds
    }
}


func checkInternetConnection(url string) bool {
    timeout := time.Duration(5 * time.Second)
    client := http.Client{
        Timeout: timeout,
    }
    _, err := client.Get(url)
    if err != nil {
        fmt.Println("Error checking internet connection:", err)
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
    // fmt.Println(data)
    if strings.HasPrefix(data, "$GNRMC") {
        if strings.Contains(data, "A") { // Check for 'Active' status
            latitude := getValue(data, ',', 3)
            ns := getValue(data, ',', 4)
            longitude := getValue(data, ',', 5)
            ew := getValue(data, ',', 6)

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
                fmt.Printf("Valid GPS Data: Latitude: %.6f, Longitude: %.6f\n", lat, lon)

                GPSCheckLED.High()

                gpsActive = true


            } else {
                fmt.Println("Invalid GPS Data: Skipping zero coordinates.")
            GPSCheckLED.Low()



            }
        } else {
            fmt.Println("Invalid GPS Data: No 'Active' status.")
            GPSCheckLED.Low()



        }
    }
}

func getValue(data string, separator rune, index int) string {
    parts := strings.Split(data, string(separator))
    if index < len(parts) {
        return parts[index]
    }
    return ""
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