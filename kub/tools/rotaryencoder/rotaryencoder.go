package main

import (
	"fmt"
	"os"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

const (
    pinA     = 17  // GPIO 17
    pinB     = 18  // GPIO 27
    button   = 22  // GPIO 22 for the switch
)

func main() {
    if err := rpio.Open(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
    defer rpio.Close()

    encoderA := rpio.Pin(pinA)
    encoderB := rpio.Pin(pinB)
    pushButton := rpio.Pin(button)

    encoderA.Input()
    encoderB.Input()
    pushButton.Input()
    pushButton.PullUp()  // Enable internal pull-up resistor

    lastValueA := encoderA.Read()
    lastValueB := encoderB.Read()


    for {

        if (encoderA.Read() != lastValueA && encoderB.Read() == lastValueB) {
            lastValueA = encoderA.Read()
            fmt.Println("Clock", lastValueA)

        } else if (encoderB.Read() != lastValueB && encoderA.Read() == lastValueA) {
            lastValueB = encoderB.Read()
            fmt.Println("Counterclock", lastValueB)
        }

        if pushButton.Read() == rpio.Low {
            fmt.Println("Button Pressed")
        }

        time.Sleep(10 * time.Millisecond)  // Adjust this delay to fine-tune performance
    }
}
