package store

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type UserScrip struct {
	UserId    string     `json:"user_id"`
	Ticker    string     `json:"ticker"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at"`
}

func (s *UserScrip) TableName() string {
	return "user_scrip"
}

type UserScripStore struct {
	db *gorm.DB
}

func (b *UserScripStore) Add(scrip UserScrip) (*UserScrip, error) {
	if err := b.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "user_id"}, {Name: "ticker"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"deleted_at": nil}),
		}).
		Select("user_id", "ticker").Create(&scrip).
		Scan(&scrip).Error; err != nil {
		return nil, err
	}
	return &scrip, nil
}

func (b *UserScripStore) Remove(scips UserScrip) error {
	return b.db.Model(&UserScrip{}).
		Where("user_id = ? AND ticker = ?", scips.UserId, scips.Ticker).
		Update("deleted_at", time.Now()).
		Error
}

func (b *UserScripStore) Get(userId string) ([]TickerInfo, error) {
	var queryResult BseScrips

	err := b.db.Model(&UserScrip{}).Select("bse_scrip.*").
		Joins("INNER JOIN bse_scrip ON user_scrip.ticker = bse_scrip.security_id").
		Where("user_scrip.user_id = ?", userId).
		Where("deleted_at is NULL").
		Scan(&queryResult).Error
	if err != nil {
		return nil, err
	}

	return queryResult.ToTickerInfos(), nil
}

func (b *UserScripStore) ListSubscribers(ticker string) ([]string, error) {

	var queryResult []UserScrip

	err := b.db.Model(&UserScrip{}).
		Select("user_scrip.*").
		Where("ticker = ?", ticker).
		Joins("INNER JOIN users ON users.id = user_scrip.user_id").
		Where("users.notifications = true").
		Where("user_scrip.deleted_at is NULL").
		Scan(&queryResult).Error
	if err != nil {
		return nil, err
	}

	var subscribers []string

	for _, i := range queryResult {
		subscribers = append(subscribers, i.UserId)
	}

	return subscribers, nil

}

func NewUserScripStore(db *gorm.DB) *UserScripStore {
	return &UserScripStore{db: db}
}
