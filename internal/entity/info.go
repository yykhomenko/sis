package entity

import "time"

type Info struct {
	Msisdn       uint64
	BillingType  uint8
	LanguageType uint8
	OperatorType uint8
	ChangeDate   time.Time
}
