package model

import (
	"time"

	"gorm.io/datatypes"
)

type Deliverable struct {
	ID               uint64         `gorm:"primaryKey" json:"id"`
	RunID            uint64         `gorm:"not null;index" json:"run_id"`
	Title            string         `gorm:"size:256;not null" json:"title"`
	Summary          string         `gorm:"type:text" json:"summary"`
	ResultJSON       datatypes.JSON `gorm:"type:json" json:"result_json"`
	ReviewerPersonID uint64         `gorm:"index" json:"reviewer_person_id"`
	ReviewStatus     string         `gorm:"size:32;not null;index" json:"review_status"`
	ReviewedAt       *time.Time     `json:"reviewed_at"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (Deliverable) TableName() string {
	return "deliverables"
}
