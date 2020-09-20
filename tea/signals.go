package tea

import (
	"os"
	"os/signal"
)

// WaitForSignal waits for syscalls (INT, TERM, ...)
func WaitForSignal(signals ...os.Signal) {
	quit := make(chan os.Signal)
	signal.Notify(quit, signals...)
	<-quit
}
