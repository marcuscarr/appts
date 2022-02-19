package models

import (
	"time"

	"gorm.io/gorm"
)

type Appt struct {
	gorm.Model
	StartTime time.Time
	EndTime   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time

	User    User
	Trainer Trainer
}

type User struct {
	gorm.Model
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time

	Appts []Appt `gorm:"constraint:ON DELETE CASCADE;"`
}

type Trainer struct {
	gorm.Model
	Name      string
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time

	Appts []Appt `gorm:"constraint:ON DELETE CASCADE;"`
}
