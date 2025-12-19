package models

import "gorm.io/gorm"

type Transaction struct {
    gorm.Model
    ProductID string  `json:"product_id" gorm:"uniqueIndex;not null;type:varchar(50)"` // Tambah type varchar(50) buat limit
    Quantity  int     `json:"quantity" gorm:"not null;check:quantity > 0"`           // Tambah check constraint
    Price     float64 `json:"price" gorm:"not null;type:decimal(10,2)"`             // Decimal buat uang, lebih akurat
    Total     float64 `json:"total" gorm:"-"`                                         // Computed
}

// Hook auto-calculate total (sama)
func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
    t.Total = float64(t.Quantity) * t.Price
    return nil
}