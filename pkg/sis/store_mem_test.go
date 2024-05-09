package sis

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testInfo = &Info{
	Msisdn:       0501234567,
	BillingType:  1,
	LanguageType: 2,
	OperatorType: 1,
}

func TestStore(t *testing.T) {
	config := NewConfig()
	store := NewStoreMem(config)
	ctx := context.Background()

	t.Run("Set", func(t *testing.T) {
		err := store.Set(ctx, testInfo)
		assert.Nil(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		info, err := store.Get(ctx, testInfo.Msisdn)
		assert.Nil(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, testInfo.Msisdn, info.Msisdn)
		assert.Equal(t, testInfo.BillingType, info.BillingType)
		assert.Equal(t, testInfo.LanguageType, info.LanguageType)
		assert.Equal(t, testInfo.OperatorType, info.OperatorType)
		assert.NotNil(t, testInfo.ChangeDate)
	})
}

func BenchmarkStore_Get(b *testing.B) {
	config := NewConfig()
	store := NewStoreMem(config)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Get(ctx, testInfo.Msisdn)
	}
}

func BenchmarkStore_Set(b *testing.B) {
	config := NewConfig()
	store := NewStoreMem(config)
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.Set(ctx, testInfo)
	}
}
