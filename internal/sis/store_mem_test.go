package sis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testSubscriber = &Subscriber{
	Msisdn:       501234567,
	BillingType:  1,
	LanguageType: 2,
	OperatorType: 1,
}

func TestStore(t *testing.T) {
	config := NewConfig()
	store := NewStoreMem(config)
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
