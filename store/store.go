package store

import (
	"errors"
	"gorm.io/gorm"
)

type Store struct {
	*BseTickerStore
	*UserStore
	*UserScripStore
	*CorporateNewsStore
	*CorporateActionStore
}

var ErrDuplicate = errors.New("duplicate item, already exists")

func NewStore(db *gorm.DB) *Store {
	return &Store{
		NewBseScripStore(db),
		NewUserStore(db),
		NewUserScripStore(db),
		NewCorporateNewsStore(db),
		NewCorporateActionStore(db),
	}
}
