package model

import "time"

type FlowTemplate struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:128;not null" json:"name"`
	Code        string    `gorm:"size:64;not null;uniqueIndex" json:"code"`
	Version     int       `gorm:"not null" json:"version"`
	Category    string    `gorm:"size:64;index" json:"category"`
	Description string    `gorm:"type:text" json:"description"`
	Status      string    `gorm:"size:32;not null;index" json:"status"`
	CreatedBy   uint64    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (FlowTemplate) TableName() string {
	return "flow_templates"
}
