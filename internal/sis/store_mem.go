package sis

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type inMemStore struct {
	config      *Config
	subscribers map[int64]*Subscriber
	sync.RWMutex
}

func NewStoreMem(config *Config) Store {
	store := &inMemStore{
		config:      config,
		subscribers: make(map[int64]*Subscriber, len(config.NDCS)*config.NDCCap),
	}

	store.Generate()

	return store
}

func (s *inMemStore) Get(ctx context.Context, msisdn int64) (*Subscriber, error) {
	s.RLock()
	subscriber, exist := s.subscribers[msisdn]
	s.RUnlock()
	if !exist {
		return nil, errors.New("subscriber not exist")
	}
	return subscriber, nil
}

func (s *inMemStore) Set(ctx context.Context, subscriber *Subscriber) error {
	subscriber.UpdatedAt = time.Now()
	s.Lock()
	s.subscribers[subscriber.Msisdn] = subscriber
	s.Unlock()
	return nil
}

func (s *inMemStore) Generate() {
	log.Printf("generate %d subscribers...", len(s.config.NDCS)*s.config.NDCCap)
	defer timeTrack(time.Now(), "generate")

	for _, ndc := range s.config.NDCS {
		s.generate(ndc)
	}
}

func (s *inMemStore) generate(ndc int) {
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
	for i := 0; i < workers; i++ {
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
					log.Println("set subscriber error: {}", err.Error())
				}
			}
		}()
	}
	wg.Wait()
}
