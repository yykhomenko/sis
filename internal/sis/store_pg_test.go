package sis

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Integration tests for the PostgreSQL store. They run only when
// SIS_TEST_DB_URL points to a throwaway database: the subscribers table
// in it is dropped and recreated.
func newTestPG(t *testing.T) (*pgStore, *pgxpool.Pool) {
	t.Helper()
	url := os.Getenv("SIS_TEST_DB_URL")
	if url == "" {
		t.Skip("SIS_TEST_DB_URL is not set")
	}

	ctx := context.Background()
	admin, err := pgxpool.New(ctx, url)
	require.NoError(t, err)
	t.Cleanup(admin.Close)

	migration, err := os.ReadFile("../../sql/schema/001_subscribers.sql")
	require.NoError(t, err)
	up := strings.Split(string(migration), "-- +goose Down")[0]

	_, err = admin.Exec(ctx, "drop table if exists subscribers")
	require.NoError(t, err)
	_, err = admin.Exec(ctx, up)
	require.NoError(t, err)

	config := newTestConfig()
	config.DbURL = url
	store := NewStorePG(config).(*pgStore)
	t.Cleanup(store.Close)
	return store, admin
}

func TestPGStore(t *testing.T) {
	store, admin := newTestPG(t)
	ctx := context.Background()

	_, err := store.Get(ctx, 380670000001)
	assert.ErrorIs(t, err, ErrNotFound)

	require.NoError(t, store.Set(ctx, &Subscriber{Msisdn: 380670000001, BillingType: 1, LanguageType: 2, OperatorType: 3}))
	got, err := store.Get(ctx, 380670000001)
	require.NoError(t, err)
	assert.Equal(t, int16(1), got.BillingType)
	assert.Equal(t, int16(2), got.LanguageType)
	assert.Equal(t, int16(3), got.OperatorType)
	assert.WithinDuration(t, time.Now(), got.UpdatedAt, time.Hour*24)

	assert.NoError(t, store.Ping(ctx))

	t.Run("Generate", func(t *testing.T) {
		require.NoError(t, store.Generate(ctx))
		var count, minN, maxN int64
		require.NoError(t, admin.QueryRow(ctx, "select count(*), min(msisdn), max(msisdn) from subscribers").Scan(&count, &minN, &maxN))
		assert.Equal(t, int64(1000), count)
		assert.Equal(t, int64(380670000000), minN)
		assert.Equal(t, int64(380670000999), maxN)

		// Running again upserts the same rows.
		require.NoError(t, store.Generate(ctx))
		require.NoError(t, admin.QueryRow(ctx, "select count(*) from subscribers").Scan(&count))
		assert.Equal(t, int64(1000), count)
	})

	t.Run("check detects wrong schema", func(t *testing.T) {
		_, err := admin.Exec(ctx, "alter table subscribers rename column updated_at to created_at")
		require.NoError(t, err)
		err = store.check()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "updated_at")
	})
}
