package sis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"time"
)

// ErrNotFound is returned by Store.Get when the subscriber does not exist.
var ErrNotFound = errors.New("subscriber not found")

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
	Ping(ctx context.Context) error
	Close()
}

// Generator is implemented by stores that can fill themselves with test data.
type Generator interface {
	Generate(ctx context.Context) error
}

// msisdnRange returns the first and last MSISDN of a block of `capacity`
// numbers for the given CC and NDC, where every MSISDN has `length` digits.
// E.g. cc="380", ndc=67, capacity=10_000_000, length=12 ->
// 380670000000..380679999999.
func msisdnRange(cc string, ndc, capacity, length int) (int, int, error) {
	ccNum, err := strconv.Atoi(cc)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid CC %q: %w", cc, err)
	}

	ndcDigits := len(strconv.Itoa(ndc))
	subscriberDigits := length - len(cc) - ndcDigits
	if subscriberDigits <= 0 {
		return 0, 0, fmt.Errorf("MSISDN length %d is too short for CC %s and NDC %d", length, cc, ndc)
	}

	ndcShift := int(math.Pow10(subscriberDigits))
	if capacity < 1 || capacity > ndcShift {
		return 0, 0, fmt.Errorf("capacity %d must be in 1..%d", capacity, ndcShift)
	}

	ccShift := int(math.Pow10(ndcDigits + subscriberDigits))

	minNum := ccNum*ccShift + ndc*ndcShift
	return minNum, minNum + capacity - 1, nil
}

func timeTrack(start time.Time, name string) {
	log.Printf("%s took %s", name, time.Since(start))
}
