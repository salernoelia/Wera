package main

import (
	"fmt"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

func main() {
	// Open the GPIO memory range for use in the program
	if err := rpio.Open(); err != nil {
		fmt.Printf("Error opening GPIO: %v\n", err)
		return
	}
	defer rpio.Close()

	// Define the GPIO pin number
	// Change the pin number based on your setup
	pin := rpio.Pin(27)
	pin.Input()       // Set the pin to input mode
	pin.PullUp()      // Activate pull up resistor

	fmt.Println("Monitoring the button... Press CTRL+C to exit.")

	// Poll the button state in an infinite loop
	for {
		// Read the pin state
		if pin.Read() == rpio.Low {
			fmt.Println("Button pressed!")
			time.Sleep(time.Second) // Add a delay to debounce the button press
		}
		time.Sleep(10 * time.Millisecond) // Polling delay to reduce CPU usage
	}
}

