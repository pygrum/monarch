package db

import (
	"github.com/pygrum/monarch/pkg/teamserver/roles"
	"time"
)

type Builder struct {
	// A UUID that identifies an agent
	BuilderID   string `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string `gorm:"unique"` // This is technically the agent name and is used as so
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
	CreatedBy string // TODO:use this field for notifications and whatnot
	CreatedAt time.Time
	AgentInfo string
}

// Profile is used to save build configurations
type Profile struct {
	ID        uint `gorm:"primaryKey"`
	UpdatedAt time.Time
	CreatedAt time.Time
	Name      string `gorm:"unique"`
	BuilderID string
	CreatedBy string
}

// ProfileRecord is one build configuration, and is bound to a profile in the profiles table
type ProfileRecord struct {
	ID        uint `gorm:"primaryKey"`
	UpdatedAt time.Time
	CreatedAt time.Time
	Profile   string
	Name      string
	Value     string
}

type Player struct {
	UUID      string `gorm:"primaryKey"`
	Username  string `gorm:"unique"`
	ClientCA  string // base64 representation of client certificate for mTLS
	Challenge string
	Secret    string
	Role      roles.Role
	CreatedAt time.Time
}
