package thuder

import (
	"errors"
	"sync"
	"time"
)

var PinID = -1 //set this use a pin

var gpioErr = errors.New("Not Initialized")

var lightOn = func() { LogP("lightOn (error: %s)\n", gpioErr) }

var lightOff = func() { LogP("lightOff (error: %s)\n", gpioErr) }

var ledCounter int
var flashing bool
var ledLock sync.Mutex

//FlashLED flashes the led at PinID once, if possible
func FlashLED() {
	ledLock.Lock()
	defer ledLock.Unlock()
	if gpioErr != nil {
		gpioErr = setupGPIO()
	}
	ledCounter++
	if !flashing {
		flashing = true
		go func() {
			ledLock.Lock()
			defer ledLock.Unlock()
			c := 1
			for ledCounter > 0 {
				if c < ledCounter {
					c = ledCounter
				}
				ledCounter--
				ledLock.Unlock()
				t := 500 * time.Microsecond / time.Duration(c)
				LogP("flash for %v\n", t)
				LEDOn()
				time.Sleep(t)
				LEDOff()
				time.Sleep(t)
				ledLock.Lock()
			}
			flashing = false
		}()
	}
}

var ledGroup sync.WaitGroup
var onCounter int

func LEDOn() {
	ledLock.Lock()
	defer ledLock.Unlock()
	if gpioErr != nil {
		gpioErr = setupGPIO()
	}
	onCounter++
	lightOn()
}

func LEDOff() {
	ledLock.Lock()
	defer ledLock.Unlock()
	onCounter--
	if onCounter == 0 {
		lightOff()
	}
}
