package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/duh-rpc/duh-go/demo"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type config struct {
	Address string
	Verbose bool
}

func checkErr(err error, format string, a ...any) {
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("%s: %s\n", format, err), a...)
		os.Exit(1)
	}
}

func fail(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, fmt.Sprintf("%s\n", format), a...)
	os.Exit(1)
}

func main() {
	var c config

	f := flag.NewFlagSet("demo", flag.ExitOnError)
	f.StringVar(&c.Address, "address", "localhost:8080",
		"The address to bind the server to in the format '<host|ip>:<port>'")
	f.BoolVar(&c.Verbose, "verbose", false,
		"be verbose")
	flag.Usage = func() {
		_, _ = fmt.Fprintf(os.Stderr, "Usage: %s [flags]\n"+
			"Flags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	checkErr(f.Parse(os.Args[1:]), "while parsing command line args")

	// Create a new instance of our service
	service := demo.NewService()

	// Support H2C (HTTP/2 ClearText)
	// See https://github.com/thrawn01/h2c-golang-example
	h2s := &http2.Server{}

	server := &http.Server{
		Handler: h2c.NewHandler(&demo.Handler{Service: service}, h2s),
		Addr:    c.Address,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Listening on %s....\n", c.Address)
		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			fail("ListenAndServe(): %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down the service...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Shutdown(ctx)
	checkErr(err, "during service shutdown")

	log.Println("Service shutdown complete")
}
