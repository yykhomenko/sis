package sis

import (
	"github.com/caarlos0/env/v11"
	"log"
)

type Config struct {
	Addr         string `env:"SIS_ADDR" envDefault:":8080"`
	CC           string `env:"SIS_CC" envDefault:"380"`
	NDCS         []int  `env:"HASHES_NDCS" envSeparator:"," envDefault:"67"`
	MsisdnLength int    `env:"SIS_MSISDN_LENGTH" envDefault:"12"`
}

func NewConfig() *Config {
	c := Config{}
	if err := env.Parse(&c); err != nil {
		log.Fatal(err)
	}
	return &c
}
