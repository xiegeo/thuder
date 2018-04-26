package thuder

import (
	"errors"
	"sync"
	"time"
)

var PinID = -1 //set this use a pin

var gpioErr = errors.New("No Hardward Support")

var lightOn = func() { LogP("lightOn (error: %s)\n", gpioErr) }

var lightOff = func() { LogP("lightOff (error: %s)\n", gpioErr) }

var ledCounter int
var flashing bool
var ledLock sync.Mutex

func flashLED() {
	ledLock.Lock()
	defer ledLock.Unlock()
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
				lightOn()
				time.Sleep(t)
				lightOff()
				time.Sleep(t)
				ledLock.Lock()
			}
			flashing = false
		}()
	}
}
