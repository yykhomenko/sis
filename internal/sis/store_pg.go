package sis

import (
	"context"
	"log"
	"math"
	"math/rand"
	"runtime"
	"strconv"
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

	if config.Generate {
		store.Generate()
	}

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

	cc, err := strconv.Atoi(s.config.CC)
	if err != nil {
		log.Fatalf("unable to parse CC, convert %s to int: %v\n", s.config.CC, err)
	}

	capacity := s.config.NDCCap

	for _, ndc := range s.config.NDCS {
		log.Printf("NDC: %d", ndc)
		s.generate(cc, ndc, capacity)
	}
}

func (s *pgStore) generate(cc int, ndc int, capacity int) {

	subscriberDigits := int(math.Log10(float64(capacity-1))) + 1
	ndcShift := int(math.Pow10(subscriberDigits))

	ndcDigits := int(math.Log10(float64(ndc))) + 1
	ccShift := int(math.Pow10(ndcDigits + subscriberDigits))

	minNum := cc*ccShift + ndc*ndcShift
	maxNum := minNum + capacity - 1

	var numWorkers = runtime.NumCPU()

	numbers := make(chan int, 1000*numWorkers)
	go func() {
		defer close(numbers)
		start := time.Now()
		for number := minNum; number <= maxNum; number++ {
			numbers <- number
			if number%(capacity/100) == 0 || number == maxNum {
				elapsed := time.Since(start).Seconds()
				processed := number - minNum
				remain := maxNum - number
				tps := float64(processed) / elapsed

				remainTime := time.Duration(float64(remain)/tps) * time.Second

				log.Printf("generate subscriber: %d/%d tps: %.0f remain: %v\n", number, maxNum-minNum+1, tps, remainTime)
			}
		}
	}()

	wg := &sync.WaitGroup{}
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(numbers <-chan int) {
			defer wg.Done()

			ctx := context.Background()
			for number := range numbers {

				subscriber := &Subscriber{
					Msisdn:       int64(number),
					BillingType:  int16(rand.Int31n(2)),
					LanguageType: int16(rand.Int31n(2)),
					OperatorType: int16(rand.Int31n(2)),
					UpdatedAt:    time.Now(),
				}

				//subscriber = subscriber

				err := s.Set(ctx, subscriber)
				if err != nil {
					log.Println("set subscriber error:", err)
				}
			}
		}(numbers)
	}
	wg.Wait()
}
