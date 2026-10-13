package models

import "time"

type MangaUsers struct {
	Id           uint64    `gorm:"primary_key;column:id" json:"id"`
	Username     string    `gorm:"column:username" json:"username"`
	Email        string    `gorm:"column:email" json:"email"`
	PasswordHash string    `gorm:"column:password_hash" json:"passwordHash"`
	CreatedDate  time.Time `gorm:"column:created_date" json:"createdDate"`
	UpdatedDate  time.Time `gorm:"column:updated_date" json:"updatedDate"`
}

func (MangaUsers) TableName() string {
	return "manga_users"
}
