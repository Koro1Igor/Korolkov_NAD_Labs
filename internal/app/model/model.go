package model

import "time"

const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusDeleted   = "deleted"
)

type User struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"type:varchar(100);not null"`
	Email     string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"not null"`
}

func (User) TableName() string { return "users" }

type LaunchVehicle struct {
	ID               uint      `gorm:"primaryKey"`
	Name             string    `gorm:"type:varchar(120);not null"`
	ShortDescription string    `gorm:"type:varchar(500)"`
	Status           string    `gorm:"type:varchar(16);not null;index"`
	ImageURL         string    `gorm:"type:varchar(500)"`
	VideoURL         string    `gorm:"type:varchar(500)"`
	PayloadKg        int       `gorm:"not null;default:0"`
	SeaLevelThrustKN int       `gorm:"not null;default:0"`
	CreatedAt        time.Time `gorm:"not null"`
	FormedAt         *time.Time
	CreatorID        uint   `gorm:"not null;index"`
	Creator          User   `gorm:"foreignKey:CreatorID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Likes            []Like `gorm:"foreignKey:LaunchVehicleID"`
}

func (LaunchVehicle) TableName() string { return "launch_vehicles" }

type Like struct {
	ID              uint          `gorm:"primaryKey"`
	UserID          uint          `gorm:"not null;uniqueIndex:idx_like_user_vehicle"`
	LaunchVehicleID uint          `gorm:"not null;uniqueIndex:idx_like_user_vehicle"`
	CreatedAt       time.Time     `gorm:"not null"`
	User            User          `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	LaunchVehicle   LaunchVehicle `gorm:"foreignKey:LaunchVehicleID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
}

func (Like) TableName() string { return "likes" }
