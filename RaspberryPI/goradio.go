package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
	"go.bug.st/serial.v1"
)

func fetchAndPlayAudio(url string) {
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

func getValue(data string, separator rune, index int) string {
	parts := strings.Split(data, string(separator))
	if index < len(parts) {
		return parts[index]
	}
	return ""
}

func convertToDecimalDegrees(coordinate string, isLatitude bool) float64 {
	degreeLength := 3
	if isLatitude {
		degreeLength = 2
	}
	var degrees, minutes float64
	fmt.Sscanf(coordinate[:degreeLength], "%f", &degrees)
	fmt.Sscanf(coordinate[degreeLength:], "%f", &minutes)
	return degrees + minutes/60
}

func processGPSData(data string) {
	if strings.HasPrefix(data, "$GNRMC") {
		if strings.Contains(data[7:], "A") {
			time := getValue(data, ',', 1)
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

			if lat == 0.0 && lon == 0.0 {
				fmt.Println("Invalid GPS Data: Skipping zero coordinates.")
				return
			}

			fmt.Printf("Valid GPS Data: Time: %s, Latitude: %.6f, Longitude: %.6f\n", time, lat, lon)
		}
	}
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
				fmt.Println("Failed to read from serial port:", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			if c == '\n' {
				if strings.HasPrefix(gpsDataBlock, "$GNRMC") || strings.HasPrefix(gpsDataBlock, "$GNGLL") {
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
			fetchAndPlayAudio("http://estationserve.ddns.net:9000/weather")
			time.Sleep(time.Second) // Debounce delay
		}
		time.Sleep(10 * time.Millisecond) // Polling delay to reduce CPU usage
	}
}
