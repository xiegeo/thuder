package thuder

import (
	"syscall"
)

func syncWriteCache() {
	syscall.Sync()
}
