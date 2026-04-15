package model

import "time"

type Person struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	Name           string    `gorm:"size:64;not null" json:"name"`
	Email          string    `gorm:"size:128;not null;index" json:"email"`
	RoleType       string    `gorm:"size:64;not null;index" json:"role_type"`
	ExternalSource *string   `gorm:"size:64;uniqueIndex:idx_person_external_identity" json:"external_source"`
	ExternalUserID *string   `gorm:"size:128;uniqueIndex:idx_person_external_identity" json:"external_user_id"`
	Status         int8      `gorm:"not null;index" json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Person) TableName() string {
	return "persons"
}
