package sis

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yykhomenko/sis/internal/database"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	pgMaxConns        = 64
	pgStartupTimeout  = 5 * time.Second
	generateBatchSize = 100_000
)

type pgStore struct {
	config  *Config
	pool    *pgxpool.Pool
	queries *database.Queries
}

func NewStorePG(config *Config) Store {
	poolConfig, err := pgxpool.ParseConfig(config.DbUrl)
	if err != nil {
		log.Fatalf("unable to config pool: %v\n", err)
	}

	poolConfig.MaxConns = pgMaxConns

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatalf("unable to create database pool: %v\n", err)
	}

	store := &pgStore{
		config:  config,
		pool:    pool,
		queries: database.New(pool),
	}

	if err := store.check(); err != nil {
		pool.Close()
		log.Fatalf("database is not ready: %v\n", err)
	}

	log.Printf("connected to database: %s:%d/%s\n",
		poolConfig.ConnConfig.Host,
		poolConfig.ConnConfig.Port,
		poolConfig.ConnConfig.Database,
	)

	return store
}

// check makes sure the database is reachable and the subscribers table has
// the schema the queries expect, so misconfiguration fails at startup instead
// of turning every request into a 500.
func (s *pgStore) check() error {
	ctx, cancel := context.WithTimeout(context.Background(), pgStartupTimeout)
	defer cancel()

	if err := s.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	if _, err := s.queries.GetSubscriber(ctx, 0); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("subscribers table: %w", err)
	}

	return nil
}

func (s *pgStore) Get(ctx context.Context, msisdn int64) (*Subscriber, error) {
	row, err := s.queries.GetSubscriber(ctx, msisdn)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &Subscriber{
		Msisdn:       row.Msisdn,
		UpdatedAt:    row.UpdatedAt.Time,
		BillingType:  row.BillingType,
		LanguageType: row.LanguageType,
		OperatorType: row.OperatorType,
	}, nil
}

func (s *pgStore) Set(ctx context.Context, subscriber *Subscriber) error {
	return s.queries.UpdateSubscriber(ctx, database.UpdateSubscriberParams{
		Msisdn:       subscriber.Msisdn,
		BillingType:  subscriber.BillingType,
		LanguageType: subscriber.LanguageType,
		OperatorType: subscriber.OperatorType,
	})
}

func (s *pgStore) Close() {
	s.pool.Close()
}

// Generate fills the subscribers table with random test data for every
// configured NDC. It is meant to run in the background and stops when ctx
// is cancelled.
func (s *pgStore) Generate(ctx context.Context) error {
	log.Printf("generate %d subscribers...", len(s.config.NDCS)*s.config.NDCCap)
	defer timeTrack(time.Now(), "generate")

	for _, ndc := range s.config.NDCS {
		log.Printf("NDC: %d", ndc)
		if err := s.generate(ctx, ndc); err != nil {
			return fmt.Errorf("NDC %d: %w", ndc, err)
		}
	}
	return nil
}

type numberRange struct {
	from, to int // inclusive
}

func (s *pgStore) generate(ctx context.Context, ndc int) error {
	minNum, maxNum, err := msisdnRange(s.config.CC, ndc, s.config.NDCCap, s.config.MsisdnLength)
	if err != nil {
		return err
	}
	total := maxNum - minNum + 1

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	batches := make(chan numberRange)
	go func() {
		defer close(batches)
		for from := minNum; from <= maxNum; from += generateBatchSize {
			to := min(from+generateBatchSize-1, maxNum)
			select {
			case batches <- numberRange{from, to}:
			case <-ctx.Done():
				return
			}
		}
	}()

	var (
		done     atomic.Int64
		firstErr error
		errOnce  sync.Once
		wg       sync.WaitGroup
		start    = time.Now()
	)

	workers := min(runtime.NumCPU(), pgMaxConns/2)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range batches {
				if err := s.copyBatch(ctx, batch); err != nil {
					errOnce.Do(func() {
						firstErr = err
						cancel()
					})
					return
				}

				processed := done.Add(int64(batch.to - batch.from + 1))
				tps := float64(processed) / time.Since(start).Seconds()
				remain := time.Duration(float64(int64(total)-processed)/tps) * time.Second
				log.Printf("generate subscribers: %d/%d tps: %.0f remain: %v\n", processed, total, tps, remain)
			}
		}()
	}
	wg.Wait()

	if firstErr != nil {
		return firstErr
	}
	return ctx.Err()
}

// copyBatch loads a batch of random subscribers with COPY into a temporary
// table and upserts them into subscribers in a single statement.
func (s *pgStore) copyBatch(ctx context.Context, batch numberRange) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after commit

	if _, err := tx.Exec(ctx, `create temp table tmp_subscribers (like subscribers including defaults) on commit drop`); err != nil {
		return fmt.Errorf("create temp table: %w", err)
	}

	n := batch.to - batch.from + 1
	rows := pgx.CopyFromSlice(n, func(i int) ([]any, error) {
		return []any{
			int64(batch.from + i),
			int16(rand.Int31n(2)),
			int16(rand.Int31n(2)),
			int16(rand.Int31n(2)),
		}, nil
	})

	columns := []string{"msisdn", "billing_type", "language_type", "operator_type"}
	if _, err := tx.CopyFrom(ctx, pgx.Identifier{"tmp_subscribers"}, columns, rows); err != nil {
		return fmt.Errorf("copy: %w", err)
	}

	if _, err := tx.Exec(ctx, `
		insert into subscribers (msisdn, billing_type, language_type, operator_type)
		select msisdn, billing_type, language_type, operator_type from tmp_subscribers
		on conflict (msisdn) do update
		set
			updated_at = now(),
			billing_type = excluded.billing_type,
			language_type = excluded.language_type,
			operator_type = excluded.operator_type`); err != nil {
		return fmt.Errorf("upsert: %w", err)
	}

	return tx.Commit(ctx)
}
