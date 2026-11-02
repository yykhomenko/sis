package sis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var testSubscriber = &Subscriber{
	Msisdn:       380671234567,
	BillingType:  1,
	LanguageType: 2,
	OperatorType: 1,
}

// newTestConfig returns a small config so tests don't generate millions of subscribers.
func newTestConfig() *Config {
	return &Config{
		Addr:         ":0",
		CC:           "380",
		NDCS:         []int{67},
		NDCCap:       1000,
		MsisdnLength: 12,
		DbMaxConns:   4,
		DbTimeout:    time.Second,
	}
}

func TestStore(t *testing.T) {
	store := NewStoreMem(newTestConfig())
	ctx := context.Background()

	t.Run("Set", func(t *testing.T) {
		err := store.Set(ctx, testSubscriber)
		assert.Nil(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		subscriber, err := store.Get(ctx, testSubscriber.Msisdn)
		assert.Nil(t, err)
		assert.NotNil(t, subscriber)
		assert.Equal(t, testSubscriber.Msisdn, subscriber.Msisdn)
		assert.Equal(t, testSubscriber.BillingType, subscriber.BillingType)
		assert.Equal(t, testSubscriber.LanguageType, subscriber.LanguageType)
		assert.Equal(t, testSubscriber.OperatorType, subscriber.OperatorType)
		assert.NotNil(t, testSubscriber.UpdatedAt)
	})

	t.Run("GetNotFound", func(t *testing.T) {
		subscriber, err := store.Get(ctx, 380679999999)
		assert.ErrorIs(t, err, ErrNotFound)
		assert.Nil(t, subscriber)
	})
}

func BenchmarkStore_Get(b *testing.B) {
	config := NewConfig()
	store := NewStoreMem(config)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Get(ctx, testSubscriber.Msisdn)
	}
}

func BenchmarkStore_Set(b *testing.B) {
	config := NewConfig()
	store := NewStoreMem(config)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.Set(ctx, testSubscriber)
	}
}

func TestMsisdnRange(t *testing.T) {
	minNum, maxNum, err := msisdnRange("380", 67, 10_000_000, 12)
	assert.NoError(t, err)
	assert.Equal(t, 380670000000, minNum)
	assert.Equal(t, 380679999999, maxNum)

	minNum, maxNum, err = msisdnRange("380", 67, 1000, 12)
	assert.NoError(t, err)
	assert.Equal(t, 380670000000, minNum)
	assert.Equal(t, 380670000999, maxNum)

	_, _, err = msisdnRange("380", 67, 100_000_000, 12)
	assert.Error(t, err)
}
