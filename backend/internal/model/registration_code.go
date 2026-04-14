package model

import "time"

type RegistrationCode struct {
	ID            uint64    `gorm:"primaryKey" json:"id"`
	Code          string    `gorm:"size:50;uniqueIndex;not null" json:"code"`
	Status        string    `gorm:"size:20;not null;default:active;index" json:"status"`
	IntegrationID *uint64   `gorm:"index" json:"integration_id"`
	ExpiresAt     time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (RegistrationCode) TableName() string {
	return "registration_codes"
}
