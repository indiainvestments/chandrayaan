package store

import (
	"errors"
	"gorm.io/gorm"
	"time"
)

type User struct {
	Id            string     `json:"id"`
	Notifications bool       `json:"notifications"`
	CreatedAt     *time.Time `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
	DeletedAt     *time.Time `json:"deleted_at"`
}

func (s *User) TableName() string {
	return "users"
}

type UserStore struct {
	db *gorm.DB
}

func (b *UserStore) Add(user *User) (*User, error) {
	err := b.db.Select("id").Create(user).Scan(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}
func (b *UserStore) Get(id string) (*User, error) {
	var user User
	result := b.db.Model(&User{}).Where("id = ?", id).Scan(&user)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("no record found")
	}

	return &user, nil
}

func (b *UserStore) GetOrCreate(id string) (*User, error) {
	user, err := b.Get(id)
	if err != nil {
		user, err = b.Add(&User{
			Id:            id,
			Notifications: false,
		})
		if err != nil {
			return nil, err
		}
	}

	return user, nil
}

func (b *UserStore) Delete(user *User) error {
	return errors.New("not implemented")
}

func (b *UserStore) Update(user *User) (*User, error) {

	err := b.db.Model(&User{}).
		Where("id = ?", user.Id).
		Update("notifications", user.Notifications).
		Error

	if err != nil {
		return nil, err
	}

	return b.Get(user.Id)
}

func NewUserStore(db *gorm.DB) *UserStore {
	return &UserStore{db: db}
}
