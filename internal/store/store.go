package store

import (
	"errors"
	"time"

	"github.com/yykhomenko/sis/internal/entity"
)

var ErrNotExist = errors.New("info not exist")

func NewStore() Store {
	return &inMemStore{
		infos: make(map[int64]*entity.Info),
	}
}

type Store interface {
	Get(int64) (*entity.Info, error)
	Set(info *entity.Info) error
}

type inMemStore struct {
	infos map[int64]*entity.Info
}

func (s *inMemStore) Get(msisdn int64) (*entity.Info, error) {
	info, exist := s.infos[msisdn]
	if !exist {
		return nil, ErrNotExist
	}
	return info, nil
}

func (s *inMemStore) Set(info *entity.Info) error {
	info.ChangeDate = time.Now()
	s.infos[info.Msisdn] = info
	return nil
}
