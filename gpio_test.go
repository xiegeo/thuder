package thuder

import (
	"testing"
	"time"
)

func TestGPIO(t *testing.T) {
	lightOn()
	lightOff()
	flashLED()
	PinID = 0
	time.Sleep(time.Second / 10)
	flashLED()
	flashLED()
	time.Sleep(2 * time.Second)
}
