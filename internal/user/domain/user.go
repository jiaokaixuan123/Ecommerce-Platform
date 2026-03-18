package domain

// 用户实体类型

import "time"

// domain层：服务的核心实体类型结构体
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Email     string    `gorm:"uniqueIndex;size:100" json:"email"`
	Phone     string    `gorm:"uniqueIndex;size:20" json:"phone"`
	Nickname  string    `gorm:"size:50" json:"nickname"`
	Avatar    string    `gorm:"size:255" json:"avatar"`
	Status    int8      `gorm:"default:1" json:"status"` // 1:正常 0:禁用
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// domain层：指定 GORM 映射的数据库表名为 users
func (User) TableName() string {
	return "users"
}
