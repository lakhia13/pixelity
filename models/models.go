package models

import (
	"gorm.io/gorm"
	"time"
)

// User represents a user in the system
type User struct {
	gorm.Model
	Username       string `gorm:"uniqueIndex;not null"`
	Email          string `gorm:"uniqueIndex;not null"`
	Password       string `gorm:"not null"`
	Albums         []Album
	IsActive       bool      `gorm:"default:true"`
	LastLoginAt    time.Time
	ProfilePicture string
}

// MediaItem represents a media file in the system
type MediaItem struct {
	gorm.Model
	Filename      string `gorm:"not null"`
	Path          string `gorm:"not null"`
	ThumbnailPath string
	Size          int64
	MimeType      string
	Type          string
	UserID        uint
	User          User
	Albums        []Album `gorm:"many2many:album_media_items;"`
	Metadata      string
}

// Album represents a collection of media items
type Album struct {
	gorm.Model
	Name        string `gorm:"not null"`
	Description string
	UserID      uint
	User        User
	MediaItems  []MediaItem `gorm:"many2many:album_media_items;"`
	Cover       string
	MediaCount  int `gorm:"-"`
}

// AlbumMediaItem represents the join table between albums and media items
type AlbumMediaItem struct {
	AlbumID      uint      `gorm:"primaryKey"`
	MediaItemID  uint      `gorm:"primaryKey"`
	AddedAt      time.Time `gorm:"autoCreateTime"`
	DisplayOrder int
}