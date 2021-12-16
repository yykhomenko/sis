package store

import "github.com/yykhomenko/sis/internal/entity"

type Store interface {
	Get(int64) (*entity.Info, error)
	Set(info entity.Info) error
}
