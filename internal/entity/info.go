package entity

import "time"

type Info struct {
	Msisdn       int64
	BillingType  int8
	LanguageType int8
	OperatorType int8
	ChangeDate   time.Time
}
