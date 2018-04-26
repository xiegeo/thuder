// +build !linux

package thuder

import (
	"errors"
)

func setupGPIO() error {
	return errors.New("Not Supported")
}
