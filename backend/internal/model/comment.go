package model

import "time"

type Comment struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	TargetType     string    `gorm:"size:32;not null;index:idx_comment_target" json:"target_type"`
	TargetID       uint64    `gorm:"not null;index:idx_comment_target" json:"target_id"`
	AuthorPersonID uint64    `gorm:"not null;index" json:"author_person_id"`
	Content        string    `gorm:"type:text;not null" json:"content"`
	IsResolved     bool      `gorm:"not null;default:false" json:"is_resolved"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Comment) TableName() string {
	return "comments"
}
