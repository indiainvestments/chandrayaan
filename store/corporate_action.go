package store

import (
	"gorm.io/gorm"
	"strings"
	"time"
)

type CorporateActionGetOpt struct {
	Ticker string
	Limit  int
}

type CorporateAction struct {
	ExDate      *time.Time `json:"ex_date"`
	Purpose     string     `json:"purpose"`
	Details     string     `json:"details"`
	PaymentDate *time.Time `json:"payment_date"`
	Id          string     `json:"id,omitempty"`
	Ticker      string     `json:"ticker"`
}

func (s *CorporateAction) TableName() string {
	return "corporate_action"
}

type CorporateActionStore struct {
	db *gorm.DB
}

func (b CorporateActionStore) Add(action *CorporateAction) (*CorporateAction, error) {
	err := b.db.Select("ex_date", "purpose", "details", "payment_date", "ticker").Create(action).Scan(action).Error
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return nil, ErrDuplicate
		}
	}
	return action, nil
}

func (b CorporateActionStore) Get(opt CorporateActionGetOpt) ([]CorporateAction, error) {

	if opt.Limit == 0 {
		opt.Limit = 10
	}

	var news []CorporateAction
	query := b.db.Model(&CorporateAction{}).Where("ticker = ?", opt.Ticker)
	query = query.Limit(opt.Limit)

	err := query.Scan(&news).Error
	if err != nil {
		return nil, err
	}
	return news, nil
}

func NewCorporateActionStore(db *gorm.DB) *CorporateActionStore {
	return &CorporateActionStore{db: db}
}
