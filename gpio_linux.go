package thuder

import (
	"github.com/stianeikeland/go-rpio"
)

func init() {
	gpioErr = setupGPIO()
}

func setupGPIO() error {
	err := rpio.Open()
	if err != nil {
		return err
	}
	pin := rpio.Pin(0)
	pin.Output()
	lightOn = pin.High
	lightOff = pin.Low
}
