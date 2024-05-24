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

func fetchAndPlayAudio(url string, lat, lon float64) {
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

func sendGPSData(lat, lon float64) {
    url := "http://estationserve.ddns.net:9000/weathergps"
    payload := GPSPayload{
        DeviceID:  "Device_1",
        Latitude:  lat,
        Longitude: lon,
    }
    jsonData, err := json.Marshal(payload)
    if err != nil {
        fmt.Println("Error marshaling JSON:", err)
        return
    }

    resp, err := http.Post(url, "application/json", bytes.NewReader(jsonData))
    if err != nil {
        fmt.Printf("Failed to send POST request: %v\n", err)
        return
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        fmt.Printf("Failed to read response body: %v\n", err)
        return
    }

    fmt.Printf("Server response: %s\n", string(body))
}

func main() {
    if err := rpio.Open(); err != nil {
        fmt.Printf("Error opening GPIO: %v\n", err)
        return
    }
    defer rpio.Close()

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

    pin := rpio.Pin(17)
    pin.Input()
    pin.PullUp()

    for {
        if pin.Read() == rpio.Low {
            fmt.Println("Button pressed, fetching and playing audio...")
            fetchAndPlayAudio("http://estationserve.ddns.net:9000/weathergps", lastValidLat, lastValidLon)
            time.Sleep(time.Second) // Debounce delay
        }
        time.Sleep(10 * time.Millisecond) // Polling delay to reduce CPU usage
    }
}

func processGPSData(data string) {
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
            } else {
                fmt.Println("Invalid GPS Data: Skipping zero coordinates.")
            }
        } else {
            fmt.Println("Invalid GPS Data: No 'Active' status.")
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
