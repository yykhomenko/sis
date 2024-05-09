package main

import (
	"github.com/yykhomenko/sis/pkg/sis"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	config := sis.NewConfig()

	store := sis.NewStorePG(config)

	server := sis.NewServer(config, store)
	server.Start()
}
