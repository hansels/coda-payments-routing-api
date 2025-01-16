package main

import (
	"github.com/hansels/coda-payments-routing-api/src/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	os.Exit(Main())
}

func Main() int {
	log.Println("Starting Load Balancer server...")

	// Create new server
	api := server.New(&server.Opts{ListenAddress: ":8080"})

	go api.Run()

	term := make(chan os.Signal)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	select {
	case s := <-term:
		log.Println("Exiting gracefully...", s)
	}

	log.Println("👋")
	return 0
}
