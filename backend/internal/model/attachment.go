package model

import "time"

type Attachment struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	TargetType string    `gorm:"size:32;not null;index:idx_attachment_target" json:"target_type"`
	TargetID   uint64    `gorm:"not null;index:idx_attachment_target" json:"target_id"`
	FileName   string    `gorm:"size:256;not null" json:"file_name"`
	FileURL    string    `gorm:"size:512;not null" json:"file_url"`
	FileSize   int64     `json:"file_size"`
	FileType   string    `gorm:"size:64" json:"file_type"`
	UploadedBy uint64    `gorm:"not null;index" json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
}

func (Attachment) TableName() string {
	return "attachments"
}
