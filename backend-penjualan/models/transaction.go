package models

import "time"

type Transaction struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	NamaPembeli  string    `gorm:"size:100;not null" json:"nama_pembeli"`
	ProductID    uint      `gorm:"not null" json:"product_id"`
	Product      Product   `gorm:"foreignKey:ProductID" json:"product"`
	Quantity     uint      `gorm:"not null" json:"quantity"`
	Harga        float64   `gorm:"type:numeric(15,2);not null" json:"harga"`
	Total        float64   `gorm:"type:numeric(15,2);not null" json:"total"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}