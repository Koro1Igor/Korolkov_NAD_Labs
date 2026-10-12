package model

import "time"

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusDeleted   = "deleted"
)

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
	Status           string  `gorm:"type:varchar(16);not null;index"`
	ShortDescription *string `gorm:"type:varchar(500)"`
	ImageURL         *string `gorm:"type:varchar(500)"`
	VideoURL         *string `gorm:"type:varchar(500)"`
	PayloadKg        *int
	SeaLevelThrustKN *int
	CreatedAt        time.Time `gorm:"not null"`
	FormedAt         *time.Time
	Creator          User   `gorm:"foreignKey:CreatorID;constraint:OnDelete:RESTRICT;"`
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
