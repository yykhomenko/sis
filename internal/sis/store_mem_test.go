package sis

import (
	"context"
	"fmt"
	"math"
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

func Test_my(b *testing.T) {
	generate(380, 67, 1_000)
}

func generate(cc int, ndc int, cap int) {

	subscriberDigits := int(math.Log10(float64(cap-1))) + 1
	ndcShift := int(math.Pow10(subscriberDigits))

	ndcDigits := int(math.Log10(float64(ndc))) + 1
	ccShift := int(math.Pow10(ndcDigits + subscriberDigits))

	minNum := cc*ccShift + ndc*ndcShift
	maxNum := minNum + cap - 1

	fmt.Println(minNum, maxNum)
}
