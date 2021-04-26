package store

import (
	"gorm.io/gorm"
)

type BseTicker struct {
	SecurityCode  string
	SecurityId    string
	SecurityName  string
	Status        string
	SecurityGroup string
	FaceValue     string
	IsinNumber    string
	Industry      string
	SecurityType  string
}

type TickerInfo struct {
	Isin     string `json:"isin,omitempty"`
	Name     string `json:"name,omitempty"`
	Ticker   string `json:"ticker,omitempty"`
	Code     string `json:"code,omitempty"`
	Industry string `json:"industry,omitempty"`
}

type BseScrips []BseTicker

func (scrips BseScrips) ToTickerInfos() []TickerInfo {

	var result []TickerInfo
	for _, i := range scrips {
		result = append(result, TickerInfo{
			Isin:     i.IsinNumber,
			Name:     i.SecurityName,
			Ticker:   i.SecurityId,
			Code:     i.SecurityCode,
			Industry: i.Industry,
		})
	}
	return result
}

func (s *BseTicker) TableName() string {
	return "bse_scrip"
}

type BseTickerStore struct {
	db *gorm.DB
}

func (b *BseTickerStore) Search(text string) ([]TickerInfo, error) {
	searchText := text

	var databaseResult BseScrips

	err := b.db.Model(&BseTicker{}).
		Select("*, (similarity(security_id, ?) + similarity(UPPER(security_name), UPPER(?))) as sim", searchText, searchText).
		Where("security_id % ? OR UPPER(security_name) % UPPER(?)", searchText, searchText).
		Order("sim desc").
		Order("security_id").
		Where("status = 'Active'").
		Limit(10).
		Scan(&databaseResult).Error

	if err != nil {
		return nil, err
	}
	return databaseResult.ToTickerInfos(), nil

}

func (b *BseTickerStore) List() ([]TickerInfo, error) {

	var databaseResult BseScrips

	err := b.db.Model(&BseTicker{}).
		Where("status = 'Active'").
		Order("security_id").
		Scan(&databaseResult).Error

	if err != nil {
		return nil, err
	}
	return databaseResult.ToTickerInfos(), nil

}

func NewBseScripStore(db *gorm.DB) *BseTickerStore {
	return &BseTickerStore{db: db}
}
