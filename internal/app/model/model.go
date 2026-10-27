package model

import (
	"errors"
	"io"
	"sync"
	"time"
)

const (
	StatusDraft                = "draft"
	StatusPublished            = "published"
	StatusDeleted              = "deleted"
	CurrentCreatorIDValue uint = 1
)

var (
	ErrNotFound      = errors.New("not found")
	ErrDraftExists   = errors.New("draft already exists")
	ErrDraftNotFound = errors.New("draft not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrEmailExists   = errors.New("email already exists")
	currentCreatorID uint
	creatorOnce      sync.Once
)

func CurrentCreatorID() uint {
	creatorOnce.Do(func() {
		currentCreatorID = CurrentCreatorIDValue
	})
	return currentCreatorID
}

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"type:varchar(100);not null"`
	Email    string `gorm:"type:varchar(255);not null;uniqueIndex"`
	Password string `gorm:"type:varchar(255);not null"`
}

func (User) TableName() string { return "users" }

type LaunchVehicle struct {
	ID               uint    `gorm:"primaryKey"`
	Name             string  `gorm:"type:varchar(120);not null"`
	CreatorID        uint    `gorm:"not null;index"`
	ModeratorID      *uint   `gorm:"index"`
	Status           string  `gorm:"type:varchar(16);not null;index"`
	ShortDescription *string `gorm:"type:varchar(500)"`
	ImageName        *string `gorm:"type:varchar(255)"`
	VideoName        *string `gorm:"type:varchar(255)"`
	PayloadKg        *int
	SeaLevelThrustKN *int
	CreatedAt        time.Time `gorm:"not null;autoCreateTime"`
	FormedAt         *time.Time
	CompletedAt      *time.Time
	Creator          User   `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT;"`
	Moderator        *User  `gorm:"foreignKey:ModeratorID;constraint:OnDelete:RESTRICT;"`
	Likes            []Like `gorm:"foreignKey:LaunchVehicleID"`
}

func (LaunchVehicle) TableName() string { return "launch_vehicles" }

type Like struct {
	ID              uint          `gorm:"primaryKey"`
	UserID          uint          `gorm:"not null;uniqueIndex:idx_like_user_vehicle"`
	LaunchVehicleID uint          `gorm:"not null;uniqueIndex:idx_like_user_vehicle"`
	User            User          `gorm:"foreignKey:UserID;constraint:OnDelete:RESTRICT;"`
	LaunchVehicle   LaunchVehicle `gorm:"foreignKey:LaunchVehicleID;constraint:OnDelete:RESTRICT;"`
}

func (Like) TableName() string { return "likes" }

type FileInput struct {
	Name        string
	ContentType string
	Size        int64
	Reader      io.Reader
}

type CreateLaunchVehicleInput struct {
	Name  string
	Image FileInput
	Video FileInput
}

type PublishLaunchVehicleInput struct {
	ShortDescription string `json:"short_description"`
	PayloadKg        int    `json:"payload_kg"`
	SeaLevelThrustKN int    `json:"sea_level_thrust_kn"`
}

type LikeInput struct {
	Value int `json:"value"`
}

type RegisterUserInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LikeResult struct {
	LaunchVehicleID uint
	Liked           int
	LikeCount       int64
}
