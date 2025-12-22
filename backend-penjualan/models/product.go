package models

import "time"

type Product struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Nama      string    `gorm:"size:100;not null" json:"nama"`
	Harga     float64   `gorm:"type:numeric(15,2);not null" json:"harga"` // FIXED: (15,2) biar max triliunan
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"`
}