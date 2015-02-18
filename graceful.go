package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func newServer() *http.Server {
	return &http.Server{
		Addr:      ":8080",
		ConnState: connCounter,
	}
}

var defaultServer = newServer()

var wg sync.WaitGroup

// Increment WaitGroup on new connections, decrement on closed connections.
func connCounter(c net.Conn, s http.ConnState) {
	if s == http.StateNew {
		wg.Add(1)
	} else if s == http.StateClosed {
		wg.Done()
	}
}

// Disable keepalives and wait for remaining connections to go into closed state.
func shutdown() {
	log.Println("Disabling keepalive")
	defaultServer.SetKeepAlivesEnabled(false)
	log.Println("Waiting for ongoing connections to finish")
	wg.Wait()
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		<-c
		log.Println("Got SIGINT, trying to shutdown gracefully")
		shutdown()
		log.Println("Graceful shutdown complete")
		os.Exit(0)
	}()

	log.Println("Starting http")
	log.Fatal(defaultServer.ListenAndServe())
}
