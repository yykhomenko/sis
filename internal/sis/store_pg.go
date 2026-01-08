package sis

import (
	"context"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/yykhomenko/sis/internal/database"

	"github.com/jackc/pgx/v5/pgxpool"
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

	poolConfig.MaxConns = 64

	pool, err := pgxpool.NewWithConfig(
		context.Background(),
		poolConfig,
	)
	if err != nil {
		log.Fatalf("unable to connect to database: %v\n", err)
	}

	log.Printf("connected to database: %s:%d/%s\n",
		pool.Config().ConnConfig.Host,
		pool.Config().ConnConfig.Port,
		pool.Config().ConnConfig.Database,
	)

	store := &pgStore{
		config:  config,
		pool:    pool,
		queries: database.New(pool),
	}

	//store.Generate()

	return store
}

func (s *pgStore) Get(ctx context.Context, msisdn int64) (*Subscriber, error) {
	row, err := s.queries.GetSubscriber(ctx, msisdn)
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

func (s *pgStore) close() {
	s.pool.Close()
}

func (s *pgStore) Generate() {
	log.Printf("generate %d subscribers...", len(s.config.NDCS)*s.config.NDCCap)
	defer timeTrack(time.Now(), "generate")

	for _, ndc := range s.config.NDCS {
		s.generate(ndc)
	}
}

func (s *pgStore) generate(ndc int) {
	minNum := ndc*10000000 + 0
	maxNum := minNum + s.config.NDCCap - 1

	var workers = runtime.GOMAXPROCS(-1)
	numbers := make(chan int, 10*workers)
	go func() {
		defer close(numbers)
		for number := minNum; number <= maxNum; number++ {
			numbers <- number
		}
	}()

	wg := &sync.WaitGroup{}
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			for number := range numbers {

				subscriber := &Subscriber{
					Msisdn:       int64(number + 380000000000),
					BillingType:  int16(rand.Int31n(2)),
					LanguageType: int16(rand.Int31n(2)),
					OperatorType: int16(rand.Int31n(2)),
					UpdatedAt:    time.Now(),
				}

				err := s.Set(ctx, subscriber)
				if err != nil {
					log.Println("set subscriber error:", err)
				}

				if number%10000 == 0 {
					log.Printf("generate subscriber: %d/%d\n", subscriber.Msisdn, maxNum-minNum)
				}
			}
		}()
	}
	wg.Wait()
}
