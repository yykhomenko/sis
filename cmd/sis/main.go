package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yykhomenko/sis/internal/sis"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	config := sis.NewConfig()

	store := sis.NewStorePG(config)
	server := sis.NewServer(config, store)

	defer store.Close()

	errCh := make(chan error, 1)
	go func() { errCh <- server.Start() }()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Printf("http-server error: %v\n", err)
	case sig := <-signals:
		log.Printf("received signal: %v\n", sig)
		if err := server.Stop(); err != nil {
			log.Printf("http-server stop error: %v\n", err)
		}
	}
}
