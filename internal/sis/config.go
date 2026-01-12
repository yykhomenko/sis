package sis

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	DbUrl        string `env:"SIS_DB_URL" envDefault:"postgresql://sis:XXXX@localhost:5432/sis?sslmode=disable"`
	Addr         string `env:"SIS_ADDR" envDefault:":9001"`
	CC           string `env:"SIS_CC" envDefault:"380"`
	NDCS         []int  `env:"SIS_NDCS" envSeparator:"," envDefault:"50,67,68,69,70,71"`
	NDCCap       int    `env:"SIS_NDC_CAPACITY" envDefault:"10000000"`
	MsisdnLength int    `env:"SIS_MSISDN_LENGTH" envDefault:"12"`
	Generate     bool   `env:"SIS_GENERATE" envDefault:"false"`
}

func NewConfig() *Config {
	c := Config{}
	if err := env.Parse(&c); err != nil {
		log.Fatal(err)
	}
	return &c
}
