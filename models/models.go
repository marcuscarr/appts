package models

import (
	"time"

	"gorm.io/gorm"
)

type Appt struct {
	gorm.Model
	ID uint `gorm:"primary_key;AUTO_INCREMENT"`

	StartTime time.Time `json:"start_time" validate:"required" gorm:"not null"`
	EndTime   time.Time `json:"end_time" validate:"required,gtfield=StartTime" gorm:"not null"`

	UserID    uint `json:"user_id" validate:"required" gorm:"not null,index"`
	TrainerID uint `json:"trainer_id" validate:"required" gorm:"not null,index"`
}

type User struct {
	gorm.Model
	ID uint `gorm:"primary_key;AUTO_INCREMENT"`

	Name     string `gorm:"not null"`
	Email    string `gorm:"not null"`
	Username string `gorm:"not null,unique"`

	Appts []Appt `gorm:"constraint:ON DELETE CASCADE;"`
}

type Trainer struct {
	gorm.Model
	ID uint `gorm:"primary_key;AUTO_INCREMENT"`

	Name     string `gorm:"not null"`
	Email    string `gorm:"not null"`
	Username string `gorm:"not null,unique"`

	Appts []Appt `gorm:"constraint:ON DELETE CASCADE;"`
}
