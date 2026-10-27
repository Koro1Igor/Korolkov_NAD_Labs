package serializer

import (
	"strings"
	"time"

	"orbitlab/internal/app/model"
)

type LaunchVehicleBase struct {
	ID               uint       `json:"id"`
	Name             string     `json:"name"`
	ShortDescription *string    `json:"short_description"`
	Status           string     `json:"status"`
	ImageName        *string    `json:"image_name"`
	VideoName        *string    `json:"video_name"`
	ImageURL         *string    `json:"image_url"`
	VideoURL         *string    `json:"video_url"`
	PayloadKg        *int       `json:"payload_kg"`
	SeaLevelThrustKN *int       `json:"sea_level_thrust_kn"`
	CreatedAt        time.Time  `json:"created_at"`
	FormedAt         *time.Time `json:"formed_at"`
	CompletedAt      *time.Time `json:"completed_at"`
	CreatorID        uint       `json:"creator_id"`
	ModeratorID      *uint      `json:"moderator_id"`
	LikeCount        int        `json:"like_count"`
}

type LaunchVehicleListItem struct {
	LaunchVehicleBase
	IsCreator int `json:"is_creator"`
}

type LaunchVehicleFeedItem struct {
	LaunchVehicleBase
	Liked     int `json:"liked"`
	IsCreator int `json:"is_creator"`
}

type UserResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type LikeResponse struct {
	ID        uint  `json:"id"`
	Liked     int   `json:"liked"`
	LikeCount int64 `json:"like_count"`
}

func ListItem(item model.LaunchVehicle, userID uint, endpoint, bucket string) LaunchVehicleListItem {
	flag := 0
	if item.CreatorID == userID {
		flag = 1
	}
	return LaunchVehicleListItem{LaunchVehicleBase: base(item, endpoint, bucket), IsCreator: flag}
}

func FeedItem(item model.LaunchVehicle, userID uint, endpoint, bucket string) LaunchVehicleFeedItem {
	liked := 0
	for _, like := range item.Likes {
		if like.UserID == userID {
			liked = 1
			break
		}
	}
	isCreator := 0
	if item.CreatorID == userID {
		isCreator = 1
	}
	return LaunchVehicleFeedItem{
		LaunchVehicleBase: base(item, endpoint, bucket),
		Liked:             liked,
		IsCreator:         isCreator,
	}
}

func Detail(item model.LaunchVehicle, endpoint, bucket string) LaunchVehicleBase {
	return base(item, endpoint, bucket)
}

func User(user model.User) UserResponse {
	return UserResponse{ID: user.ID, Name: user.Name, Email: user.Email}
}

func base(item model.LaunchVehicle, endpoint, bucket string) LaunchVehicleBase {
	return LaunchVehicleBase{
		ID:               item.ID,
		Name:             item.Name,
		ShortDescription: item.ShortDescription,
		Status:           item.Status,
		ImageName:        item.ImageName,
		VideoName:        item.VideoName,
		ImageURL:         objectURL(item.ImageName, endpoint, bucket),
		VideoURL:         objectURL(item.VideoName, endpoint, bucket),
		PayloadKg:        item.PayloadKg,
		SeaLevelThrustKN: item.SeaLevelThrustKN,
		CreatedAt:        item.CreatedAt,
		FormedAt:         item.FormedAt,
		CompletedAt:      item.CompletedAt,
		CreatorID:        item.CreatorID,
		ModeratorID:      item.ModeratorID,
		LikeCount:        len(item.Likes),
	}
}

func objectURL(name *string, endpoint, bucket string) *string {
	if name == nil || strings.TrimSpace(*name) == "" {
		return nil
	}
	value := strings.TrimRight(endpoint, "/") + "/" + strings.Trim(bucket, "/") + "/" + strings.TrimLeft(*name, "/")
	return &value
}
