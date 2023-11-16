package db

import (
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
	// TranslatorID is the UUID that identifies the translator used by the agent.
	// This would be in a separate container from the builder, which also allows authors to use
	// someone else's translator.
	// TranslatorID is the primary key for the translators table
	TranslatorID string
}

type Translator struct {
	TranslatorID string `gorm:"primaryKey"`
	Version      string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Name         string
	Author       string
	Url          string
	InstalledAt  string
	// Image and container that run the translator as a service
	ImageID     string
	ContainerID string
}
