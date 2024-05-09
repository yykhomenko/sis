package sis

import (
	"github.com/caarlos0/env/v11"
	"log"
)

type Config struct {
	DbUrl        string `env:"SIS_DB_URL,required"`
	Addr         string `env:"SIS_ADDR" envDefault:":8080"`
	CC           string `env:"SIS_CC" envDefault:"380"`
	NDCS         []int  `env:"SIS_NDCS" envSeparator:"," envDefault:"67"`
	NDCCap       int    `env:"SIS_NDC_CAPACITY" envDefault:"10000000"`
	MsisdnLength int    `env:"SIS_MSISDN_LENGTH" envDefault:"12"`
}

func NewConfig() *Config {
	c := Config{}
	if err := env.Parse(&c); err != nil {
		log.Fatal(err)
	}
	return &c
}
