package thuder

import (
	"errors"
	"sync"
	"time"
)

var ledLock sync.Mutex

var pinID = -1 //set this use a pin

func SetPinID(id int) {
	ledLock.Lock()
	defer ledLock.Unlock()

	pinID = id
}

var errOnGpio = errors.New("Not Initialized")

var lightOn = func() {
	if pinID != -1 {
		LogP("lightOn (error: %s)\n", errOnGpio)
	}
}

var lightOff = func() {
	if pinID != -1 {
		LogP("lightOff (error: %s)\n", errOnGpio)
	}
}

var ledCounter int
var flashing bool

//FlashLED flashes the led at PinID once, if possible
func FlashLED() {
	ledLock.Lock()
	defer ledLock.Unlock()
	if errOnGpio != nil {
		errOnGpio = setupGPIO()
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
				t := 500 * time.Millisecond / time.Duration(c)
				//LogP("flash for %v\n", t)
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

var onCounter int

func LEDOn() {
	ledLock.Lock()
	defer ledLock.Unlock()
	if errOnGpio != nil {
		errOnGpio = setupGPIO()
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
