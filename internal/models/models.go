package models

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// User represents a developer/student account on the platform
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"` // Hidden from JSON responses
	Projects  []Project      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"projects,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Project is a logical container for an application (e.g., "Foodie")
type Project struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	UserID      uint           `gorm:"index;not null" json:"userId"`
	Name        string         `gorm:"not null" json:"name"`
	Description string         `json:"description,omitempty"`
	APIKey      string         `gorm:"uniqueIndex;not null" json:"apiKey"` // Permanent key for client-side frontend apps
	Collections []Collection   `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"collections,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// Collection represents a virtual "table" within a Project
type Collection struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ProjectID uint           `gorm:"index;not null" json:"projectId"`
	Name      string         `gorm:"not null" json:"name"` // e.g., "users", "products"
	Documents []Document     `gorm:"foreignKey:CollectionID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"-"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Document represents an actual record inside a Collection using JSONB
type Document struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	CollectionID uint           `gorm:"index;not null" json:"collectionId"`
	Data         datatypes.JSON `gorm:"type:jsonb" json:"data"` // Stores the actual arbitrary JSON payload
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
}
