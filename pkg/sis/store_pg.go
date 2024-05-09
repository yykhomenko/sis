package sis

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type pgStore struct {
	config *Config
	pool   *pgxpool.Pool
}

func NewStorePG(config *Config) Store {

	poolConfig, err := pgxpool.ParseConfig(config.DbUrl)
	if err != nil {
		log.Fatalf("unable to config pool: %v\n", err)
	}

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

	store := &pgStore{config: config, pool: pool}

	//store.Generate()

	return store
}

func (s *pgStore) Get(ctx context.Context, msisdn int64) (*Info, error) {
	var info Info
	err := s.pool.QueryRow(ctx,
		`SELECT 
  		 msisdn, 
  		 billing_type, 
  		 language_type, 
  		 operator_type, 
  		 change_date 
     FROM 
       info 
     WHERE 
       msisdn = $1`,
		msisdn,
	).
		Scan(
			&info.Msisdn,
			&info.BillingType,
			&info.LanguageType,
			&info.OperatorType,
			&info.ChangeDate,
		)
	return &info, err
}

func (s *pgStore) Set(ctx context.Context, info *Info) error {
	_, err := s.pool.Exec(ctx,
		`INSERT INTO 
       info(msisdn, billing_type, language_type, operator_type) 
		 VALUES($1, $2, $3, $4) 
		 ON CONFLICT (msisdn) DO 
		 UPDATE 
		   SET 
		     billing_type = excluded.billing_type, 
		     language_type = excluded.language_type, 
		     operator_type = excluded.operator_type, 
		     change_date = now()`,
		info.Msisdn,
		info.BillingType,
		info.LanguageType,
		info.OperatorType,
	)
	return err
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
	minNum := ndc*s.config.NDCCap + 0
	maxNum := ndc*s.config.NDCCap + s.config.NDCCap - 1

	var workers = runtime.GOMAXPROCS(-1)
	numbers := make(chan int, 10*workers)
	go func() {
		defer close(numbers)
		for number := minNum; number <= maxNum; number++ {
			numbers <- number
		}
	}()

	wg := &sync.WaitGroup{}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			for number := range numbers {

				info := &Info{
					Msisdn:       int64(number + 380000000000),
					BillingType:  int8(rand.Int31n(2)),
					LanguageType: int8(rand.Int31n(2)),
					OperatorType: int8(rand.Int31n(2)),
					ChangeDate:   time.Now(),
				}

				err := s.Set(ctx, info)
				if err != nil {
					log.Println("set info error:", err)
				}
			}
		}()
	}
	wg.Wait()
}
