package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/stianeikeland/go-rpio/v4" // Import go-rpio to manage GPIO operations
)

// fetchAndPlayAudio takes a URL as input, fetches the audio from that URL, and plays it using ffplay.
// It logs relevant messages to the console regarding the status of each step.
func fetchAndPlayAudio(url string) {
	resp, err := http.Get(url) // Make an HTTP GET request to the specified URL
	if err != nil {
		fmt.Printf("Failed to fetch audio: %v\n", err) // Log any errors during fetching
		return
	}
	defer resp.Body.Close() // Ensure response body is closed after function exits

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Failed to fetch audio: %d\n", resp.StatusCode) // Check for HTTP errors
		return
	}

	audioData, err := io.ReadAll(resp.Body) // Read the entire response body
	if err != nil {
		fmt.Printf("Failed to read audio data: %v\n", err) // Log errors during reading
		return
	}

	// Create a temporary file to store the downloaded audio
	tmpFile, err := os.CreateTemp("", "audio-*.wav")
	if err != nil {
		fmt.Printf("Failed to create a temp file: %v\n", err) // Log file creation errors
		return
	}
	defer tmpFile.Close()           // Ensure the temporary file is closed after function exits
	defer os.Remove(tmpFile.Name()) // Ensure the temporary file is deleted after function exits

	if _, err := tmpFile.Write(audioData); err != nil {
		fmt.Printf("Failed to write to temp file: %v\n", err) // Log errors during file write
		return
	}

	fmt.Println("Playing audio...") // Notify user that playback is starting
	cmd := exec.Command("ffplay", "-nodisp", "-autoexit", tmpFile.Name())
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to play audio: %v\n", err) // Log any errors during audio playback
	}
}

func main() {
	if err := rpio.Open(); err != nil {
		fmt.Printf("Error opening GPIO: %v\n", err) // Log errors if GPIO cannot be opened
		return
	}
	defer rpio.Close() // Ensure GPIO resources are freed on program exit

	pin := rpio.Pin(17) // Set up GPIO pin 17
	pin.Input()         // Set pin 17 as input
	pin.PullUp()        // Enable pull-up resistor for pin 17

	// Loop indefinitely to listen for button presses
	for {
		if pin.Read() == rpio.Low { // Check if button is pressed (pin state is Low)
			fmt.Println("Button pressed, fetching and playing audio...")
			fetchAndPlayAudio("http://192.168.1.13:8080/weather")
			time.Sleep(time.Second) // Wait a second after a button press to debounce
		}
		time.Sleep(10 * time.Millisecond) // Sleep briefly to reduce CPU usage
	}
}
