package termination

import (
	"fmt"
	"os"
	"os/signal"
)

// Handle calls a function on termination of the program
func Handle(termFunc func()) {
	termChan := make(chan os.Signal)
	signal.Notify(termChan, os.Interrupt)

	sig := <-termChan
	fmt.Printf("Got %s signal. Aborting and calling term func...\n", sig)

	termFunc()
}
