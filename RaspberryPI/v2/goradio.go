package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
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

func main() {
	if err := rpio.Open(); err != nil {
		fmt.Printf("Error opening GPIO: %v\n", err)
		return
	}
	defer rpio.Close()

	pin := rpio.Pin(17)
	pin.Input()
	pin.PullUp()

	for {
		if pin.Read() == rpio.Low {
			fmt.Println("Button pressed, fetching and playing audio...")
			fetchAndPlayAudio("http://192.168.1.13:8080/weather")
			time.Sleep(time.Second) // Debounce delay
		}
		time.Sleep(10 * time.Millisecond) // Polling delay to reduce CPU usage
	}
}
