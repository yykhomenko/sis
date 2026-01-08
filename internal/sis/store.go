package sis

import (
	"context"
	"log"
	"time"
)

type Subscriber struct {
	Msisdn       int64     `json:"msisdn"`
	BillingType  int16     `json:"billing_type"`
	LanguageType int16     `json:"language_type"`
	OperatorType int16     `json:"operator_type"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Store interface {
	Get(ctx context.Context, msisdn int64) (*Subscriber, error)
	Set(ctx context.Context, subscriber *Subscriber) error
}

func timeTrack(start time.Time, name string) {
	log.Printf("%s took %s", name, time.Since(start))
}
