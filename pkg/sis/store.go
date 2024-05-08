package sis

import (
	"errors"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"
)

type Info struct {
	Msisdn       int64
	BillingType  int8
	LanguageType int8
	OperatorType int8
	ChangeDate   time.Time
}

var ErrNotExist = errors.New("info not exist")

func NewStore(config *Config) Store {
	store := &inMemStore{
		config: config,
		infos:  make(map[int64]*Info, len(config.NDCS)*config.NDCCap),
	}

	store.Generate()

	return store
}

type Store interface {
	Get(int64) (*Info, error)
	Set(info *Info) error
}

type inMemStore struct {
	config *Config
	sync.RWMutex
	infos map[int64]*Info
}

func (s *inMemStore) Get(msisdn int64) (*Info, error) {
	s.RLock()
	info, exist := s.infos[msisdn]
	s.RUnlock()
	if !exist {
		return nil, ErrNotExist
	}
	return info, nil
}

func (s *inMemStore) Set(info *Info) error {
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
			for number := range numbers {

				info := &Info{
					Msisdn:       int64(number + 380000000000),
					BillingType:  int8(rand.Int31n(2)),
					LanguageType: int8(rand.Int31n(2)),
					OperatorType: int8(rand.Int31n(2)),
					ChangeDate:   time.Now(),
				}

				err := s.Set(info)
				if err != nil {
					log.Println("set info error:", err)
				}
			}
		}()
	}
	wg.Wait()
}

func timeTrack(start time.Time, name string) {
	log.Printf("%s took %s", name, time.Since(start))
}
