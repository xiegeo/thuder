package thuder

import (
	"testing"
	"time"
)

func TestGPIO(t *testing.T) {
	lightOn()
	lightOff()
	FlashLED()
	PinID = 0
	time.Sleep(time.Second / 10)
	FlashLED()
	FlashLED()
	time.Sleep(2 * time.Second)
}
