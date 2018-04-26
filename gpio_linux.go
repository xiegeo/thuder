package thuder

import (
	"errors"

	"github.com/stianeikeland/go-rpio"
)

func init() {
	gpioErr = setupGPIO()
}

func setupGPIO() error {
	if PinID < 0 {
		return errors.New("need PinID")
	}
	err := rpio.Open()
	if err != nil {
		return err
	}
	pin := rpio.Pin(pinID)
	pin.Output()
	lightOn = pin.High
	lightOff = pin.Low
	return nil
}
