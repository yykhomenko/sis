package store

import (
	"errors"
	"time"

	"github.com/yykhomenko/sis/internal/entity"
)

var ErrNotExist = errors.New("info not exist")

type Store interface {
	Get(int64) (*entity.Info, error)
	Set(info *entity.Info) error
}

type database struct {
	infos map[int64]*entity.Info
}

func NewStore() Store {
	return &database{
		infos: make(map[int64]*entity.Info),
	}
}

func (db *database) Get(msisdn int64) (*entity.Info, error) {
	info, exist := db.infos[msisdn]
	if !exist {
		return nil, ErrNotExist
	}
	return info, nil
}

func (db *database) Set(info *entity.Info) error {
	info.ChangeDate = time.Now()
	db.infos[info.Msisdn] = info
	return nil
}
