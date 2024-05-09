package sis

import (
	"context"
	"log"
	"time"
)

type Subscriber struct {
	Msisdn       int64     `json:"msisdn"`
	BillingType  int8      `json:"billing_type"`
	LanguageType int8      `json:"language_type"`
	OperatorType int8      `json:"operator_type"`
	ChangeDate   time.Time `json:"change_date"`
}

type Store interface {
	Get(ctx context.Context, msisdn int64) (*Subscriber, error)
	Set(ctx context.Context, info *Subscriber) error
}

func timeTrack(start time.Time, name string) {
	log.Printf("%s took %s", name, time.Since(start))
}
