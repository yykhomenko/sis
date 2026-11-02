package main

import (
	"context"
	"errors"
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
	defer store.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // runs before store.Close, so generation stops first

	if config.Generate {
		if g, ok := store.(sis.Generator); ok {
			go func() {
				if err := g.Generate(ctx); err != nil && !errors.Is(err, context.Canceled) {
					log.Printf("generate error: %v\n", err)
				}
			}()
		}
	}

	server := sis.NewServer(config, store)

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
