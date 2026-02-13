package repository

import (
	"github.com/MorantHP/OURERP/internal/models"
	"gorm.io/gorm"
)

type Shop struct {
	ID          int64  `gorm:"primaryKey"`
	Name        string `gorm:"size:100"`
	Platform    string `gorm:"size:20"` // taobao, jd, etc.
	AppKey      string `gorm:"size:100"`
	AppSecret   string `gorm:"size:100"`
	AccessToken string `gorm:"size:500"`
	Status      int    `gorm:"default:1"`
}

type ShopRepository struct {
	db *gorm.DB
}

func NewShopRepository(db *gorm.DB) *ShopRepository {
	return &ShopRepository{db: db}
}

func (r *ShopRepository) FindByID(id int64) (*Shop, error) {
	var shop Shop
	err := r.db.First(&shop, id).Error
	if err != nil {
		return nil, err
	}
	return &shop, nil
}

func (r *ShopRepository) Create(shop *Shop) error {
	return r.db.Create(shop).Error
}