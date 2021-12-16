package store

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yykhomenko/sis/internal/entity"
)

var testInto = &entity.Info{
	Msisdn:       0501234567,
	BillingType:  1,
	LanguageType: 2,
	OperatorType: 1,
}

func TestStore(t *testing.T) {
	store := NewStore()

	err := store.Set(testInto)
	assert.Nil(t, err)

	info, err := store.Get(testInto.Msisdn)
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, testInto.Msisdn, info.Msisdn)
	assert.Equal(t, testInto.BillingType, info.BillingType)
	assert.Equal(t, testInto.LanguageType, info.LanguageType)
	assert.Equal(t, testInto.OperatorType, info.OperatorType)
	assert.NotNil(t, info.ChangeDate)
}
