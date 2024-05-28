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
	cmd := exec.Command("ffplay", "-nodisp", "-autoexit", tmpFile.Name())
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


      const (
    pinA     = 17  // GPIO 17
    pinB     = 27  // GPIO 27
    button   = 22  // GPIO 22 for the switch
    GPSLED = 26
    INTERNETLED =  19 
    )


    setupCloseHandler()

    GPSCheckLED = rpio.Pin(GPSLED)
    GPSCheckLED.Output()
    GPSCheckLED.Low()

    internetCheckLED = rpio.Pin(INTERNETLED)
    internetCheckLED.Output()
    internetCheckLED.Low()

    urlToCheck := "http://www.google.com"


    mode := &serial.Mode{
        BaudRate: 9600,
        Parity:   serial.NoParity,
        DataBits: 8,
        StopBits: serial.OneStopBit,
    }

    port, err := serial.Open("/dev/serial0", mode)
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

    encoderA := rpio.Pin(pinA)
    encoderB := rpio.Pin(pinB)
    pushButton := rpio.Pin(button)

    encoderA.Input()
    encoderB.Input()
    pushButton.Input()
    pushButton.PullUp()  // Enable internal pull-up resistor
    
 
    lastA := encoderA.Read()
    lastB := encoderB.Read()

  
    for {
        isConnected := checkInternetConnection(urlToCheck)
        if isConnected {
            internetCheckLED.Low()
        } else {
            internetCheckLED.High()
        }

        currentA := encoderA.Read()
        currentB := encoderB.Read()

        if currentA != lastA || currentB != lastB {
            if currentA == rpio.High && currentB != lastB {
                fmt.Println("Rotated Clockwise")
            } else if currentA == rpio.Low && currentB != lastB {
                fmt.Println("Rotated Counter-Clockwise")
            }
        }

        lastA = currentA
        lastB = currentB

        if pushButton.Read() == rpio.Low {
            fmt.Println("Button Pressed")
            if gpsActive && isConnected {
                fmt.Println("Trying to post to /weathergps...")
                postAndPlayAudio("https://spatial-interaction.onrender.com/weathergps", lastValidLat, lastValidLon)
            } else if !gpsActive && isConnected {
                fmt.Println("Trying to get to /weather since no GPS data is availible...")
                getAndPlayAudio("https://spatial-interaction.onrender.com/weather")
            } else {
                fmt.Println("Internet Status:", isConnected)
                fmt.Println("GPS Status:", gpsActive)

            }

            time.Sleep(1000 * time.Millisecond) // Button debounce delay
        }

        time.Sleep(2 * time.Millisecond) // Adjust this delay to fine-tune performance
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