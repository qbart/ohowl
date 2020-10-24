package tea

import (
	"os"
	"os/signal"
	"syscall"
)

// SysCallWait waits for syscalls (INT, TERM, ...)
func SysCallWait(signals ...os.Signal) {
	quit := make(chan os.Signal)
	signal.Notify(quit, signals...)
	<-quit
}

// SysCallWaitDefault waits for INT and TERM signals.
func SysCallWaitDefault() {
	SysCallWait(syscall.SIGINT, syscall.SIGTERM)
}
