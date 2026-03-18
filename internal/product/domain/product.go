package domain

import "time"

// Category 商品分类
type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:50;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Category) TableName() string { return "categories" }

// Product 商品实体
type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"not null" json:"price"`           // 单价（元）
	Stock       int       `gorm:"not null;default:0" json:"stock"` // 库存数量
	CategoryID  uint      `gorm:"not null" json:"category_id"`
	Status      int8      `gorm:"default:1" json:"status"` 		   // 1:上架 0:下架
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// TODO: 如果需要关联查询分类信息，在这里加上：
	// Category Category `gorm:"foreignKey:CategoryID" json:"category"`
}

func (Product) TableName() string { return "products" }
