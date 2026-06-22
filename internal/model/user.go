package model

import "time"

// User 用户
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;size:50" json:"username"`
	Password  string    `gorm:"size:200" json:"-"` // bcrypt hash, 永不序列化到 JSON
	Nickname  string    `gorm:"size:50" json:"nickname"`
	Role      string    `gorm:"size:20;default:user" json:"role"` // admin / user
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
