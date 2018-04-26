package thuder

import (
	"errors"
)

var gpioErr = errors.New("No Hardward Support")

var lightOn = func() { LogP("lightOn (error: %s)\n", gpioErr) }

var lightOff = func() { LogP("lightOff (error: %s)\n", gpioErr) }
