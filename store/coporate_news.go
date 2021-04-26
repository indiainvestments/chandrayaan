package store

import (
	"gorm.io/gorm"
	"strings"
	"time"
)

type CorporateNewsGetOpt struct {
	Ticker string
	From   *time.Time
	To     *time.Time
	Limit  int
}

type CorporateNews struct {
	Attachment string `json:"attachment,omitempty"`
	Headline   string `json:"headline,omitempty"`
	Date       string `json:"date,omitempty"`
	Category   string `json:"category,omitempty"`
	Id         string `json:"id,omitempty"`
	Ticker     string `json:"ticker"`
	NewsSub    string `json:"news_sub"`
}



func (s *CorporateNews) TableName() string {
	return "corporate_news"
}

type CorporateNewsStore struct {
	db *gorm.DB
}

func (b CorporateNewsStore) Add(corpNews *CorporateNews) (*CorporateNews, error) {
	err := b.db.Select("attachment", "headline", "date", "category", "id", "ticker", "news_sub").Create(corpNews).Scan(corpNews).Error
	if err != nil {
		if strings.Contains(err.Error(), "23505") {
			return nil, ErrDuplicate
		}
	}
	return corpNews, nil
}

func (b CorporateNewsStore) Get(opt CorporateNewsGetOpt) ([]CorporateNews, error) {

	if opt.Limit == 0 {
		opt.Limit = 10
	}

	var news []CorporateNews
	query := b.db.Model(&CorporateNews{}).Where("ticker = ?", opt.Ticker).
		Order("date desc")

	if opt.From != nil {
		query = query.Where("date >= ?", opt.From)
	}

	if opt.To != nil {
		query = query.Where("date <= ?", opt.To)
	}

	query = query.Limit(opt.Limit)

	err := query.Scan(&news).Error
	if err != nil {
		return nil, err
	}
	return news, nil
}

func NewCorporateNewsStore(db *gorm.DB) *CorporateNewsStore {
	return &CorporateNewsStore{db: db}
}
