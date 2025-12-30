package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents the shared user data structure.
// It serves as the single source of truth for API requests/responses and database persistence.
type User struct {
	ID        uint           `json:"id" bson:"_id,omitempty" gorm:"primaryKey" validate:"-"`
	Username  string         `json:"username" bson:"username" gorm:"uniqueIndex" validate:"required,min=3"`
	Email     string         `json:"email" bson:"email" gorm:"uniqueIndex" validate:"required,email"`
	Password  string         `json:"-" bson:"password" gorm:"not null" validate:"required"`
	Role      string         `json:"role" bson:"role" gorm:"default:'user'"`
	CreatedAt time.Time      `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" bson:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" bson:"-" gorm:"index"`
}
