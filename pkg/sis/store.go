package sis

import (
	"context"
	"log"
	"time"
)

type Info struct {
	Msisdn       int64
	BillingType  int8
	LanguageType int8
	OperatorType int8
	ChangeDate   time.Time
}

type Store interface {
	Get(ctx context.Context, msisdn int64) (*Info, error)
	Set(ctx context.Context, info *Info) error
}

func timeTrack(start time.Time, name string) {
	log.Printf("%s took %s", name, time.Since(start))
}
