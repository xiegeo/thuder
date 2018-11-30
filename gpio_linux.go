package thuder

import (
	"errors"

	"github.com/stianeikeland/go-rpio/v4"
)

func setupGPIO() error {
	if pinID < 0 {
		return errors.New("need to SetPinID")
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
