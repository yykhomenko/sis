package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yykhomenko/sis/internal/entity"
)

var testInfo = &entity.Info{
	Msisdn:       0501234567,
	BillingType:  1,
	LanguageType: 2,
	OperatorType: 1,
}

func TestStore(t *testing.T) {
	store := NewStore()

	t.Run("Set", func(t *testing.T) {
		err := store.Set(testInfo)
		assert.Nil(t, err)
	})

	t.Run("Get", func(t *testing.T) {
		info, err := store.Get(testInfo.Msisdn)
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
	store := NewStore()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Get(testInfo.Msisdn)
	}
}

func BenchmarkStore_Set(b *testing.B) {
	store := NewStore()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = store.Set(testInfo)
	}
}
