package model

import (
	"time"

	"gorm.io/datatypes"
)

type Agent struct {
	ID            uint64         `gorm:"primaryKey" json:"id"`
	Name          string         `gorm:"size:128;not null" json:"name"`
	Code          string         `gorm:"size:64;not null;uniqueIndex" json:"code"`
	Provider      string         `gorm:"size:64;not null" json:"provider"`
	Version       string         `gorm:"size:64;not null" json:"version"`
	OwnerPersonID uint64         `gorm:"index" json:"owner_person_id"`
	ConfigJSON    datatypes.JSON `gorm:"type:json" json:"config_json"`
	Status        int8           `gorm:"not null;index" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (Agent) TableName() string {
	return "agents"
}
