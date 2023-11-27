package db

import (
	"gorm.io/gorm"
	"time"
)

type Builder struct {
	// A UUID that identifies an agent
	BuilderID   string `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string // This is technically the agent name and is used as so
	Version     string
	Author      string
	Url         string
	SupportedOS string // Array of comma-separated operating systems
	// The path that the agent files are installed at
	InstalledAt string
	// ImageID and ContainerID identify the image / container used to build implants
	ImageID     string
	ContainerID string
}

type Agent struct {
	AgentID   string `gorm:"primaryKey"`
	Name      string
	Version   string
	OS        string
	Arch      string
	Host      string
	Port      string
	Builder   string // The builder used to build the agent
	File      string // binary file associated with agent
	CreatedAt time.Time
}

// Profile is used to save build configurations
type Profile struct {
	gorm.Model
	Name      string
	BuilderID string
}

// ProfileRecord is one build configuration, and is bound to a profile in the profiles table
type ProfileRecord struct {
	gorm.Model
	Profile string
	Name    string
	Value   string
}
