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
	config *Config
	sync.RWMutex
	infos map[int64]*Info
}

func NewStoreMem(config *Config) Store {
	store := &inMemStore{
		config: config,
		infos:  make(map[int64]*Info, len(config.NDCS)*config.NDCCap),
	}

	store.Generate()

	return store
}

func (s *inMemStore) Get(ctx context.Context, msisdn int64) (*Info, error) {
	s.RLock()
	info, exist := s.infos[msisdn]
	s.RUnlock()
	if !exist {
		return nil, errors.New("info not exist")
	}
	return info, nil
}

func (s *inMemStore) Set(ctx context.Context, info *Info) error {
	info.ChangeDate = time.Now()
	s.Lock()
	s.infos[info.Msisdn] = info
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

				info := &Info{
					Msisdn:       int64(number + 380000000000),
					BillingType:  int8(rand.Int31n(2)),
					LanguageType: int8(rand.Int31n(2)),
					OperatorType: int8(rand.Int31n(2)),
					ChangeDate:   time.Now(),
				}

				err := s.Set(ctx, info)
				if err != nil {
					log.Println("set info error: {}", err.Error())
				}
			}
		}()
	}
	wg.Wait()
}
