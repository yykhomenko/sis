package sis

import (
	"errors"
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

func NewStore() Store {
	return &inMemStore{
		infos: make(map[int64]*Info),
	}
}

type Store interface {
	Get(int64) (*Info, error)
	Set(info *Info) error
}

type inMemStore struct {
	infos map[int64]*Info
}

func (s *inMemStore) Get(msisdn int64) (*Info, error) {
	info, exist := s.infos[msisdn]
	if !exist {
		return nil, ErrNotExist
	}
	return info, nil
}

func (s *inMemStore) Set(info *Info) error {
	info.ChangeDate = time.Now()
	s.infos[info.Msisdn] = info
	return nil
}
