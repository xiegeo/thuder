package thuder

import (
	"testing"
	"time"
)

func TestGPIO(t *testing.T) {
	_ = time.Second
	lightOn()
	lightOff()
	FlashLED()
	//SetPinID(0)
	//time.Sleep(time.Second / 10)
	FlashLED()
	FlashLED()
	//time.Sleep(2 * time.Second)
}
