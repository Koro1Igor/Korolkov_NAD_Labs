package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"orbitlab/internal/app/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var ErrDraftNotFound = errors.New("черновик launch vehicle не найден")

type Repository struct {
	db    *gorm.DB
	sqlDB *sql.DB
}

func NewRepository() (*Repository, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		getEnv("PGHOST", "localhost"),
		getEnv("PGUSER", "orbit"),
		getEnv("PGPASSWORD", "orbit12345"),
		getEnv("PGDATABASE", "orbitlab"),
		getEnv("PGPORT", "5432"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.LaunchVehicle{},
		&model.Like{},
	); err != nil {
		return nil, fmt.Errorf("ошибка миграции: %w", err)
	}

	if err := db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_one_draft_per_creator
		ON launch_vehicles (creator_id)
		WHERE status = 'draft'
	`).Error; err != nil {
		return nil, fmt.Errorf("ошибка создания ограничения на draft: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("ошибка получения database/sql: %w", err)
	}

	return &Repository{
		db:    db,
		sqlDB: sqlDB,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (r *Repository) GetFeedLaunchVehicle(id *uint, next bool) (model.LaunchVehicle, error) {
	var launchVehicle model.LaunchVehicle

	query := r.db.
		Model(&model.LaunchVehicle{}).
		Preload("Likes").
		Where("status = ?", model.StatusPublished)

	if id == nil {
		err := query.
			Order("id ASC").
			Limit(1).
			First(&launchVehicle).Error
		return launchVehicle, err
	}

	if !next {
		err := query.
			Where("id = ?", *id).
			Limit(1).
			First(&launchVehicle).Error
		return launchVehicle, err
	}

	err := query.
		Where("id > ?", *id).
		Order("id ASC").
		Limit(1).
		First(&launchVehicle).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.
			Model(&model.LaunchVehicle{}).
			Preload("Likes").
			Where("status = ?", model.StatusPublished).
			Order("id ASC").
			Limit(1).
			First(&launchVehicle).Error
	}

	return launchVehicle, err
}

func (r *Repository) GetDraftByCreator(creatorID uint) (model.LaunchVehicle, error) {
	var launchVehicle model.LaunchVehicle

	err := r.db.
		Where(
			"creator_id = ? AND status = ?",
			creatorID,
			model.StatusDraft,
		).
		Limit(1).
		First(&launchVehicle).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return launchVehicle, ErrDraftNotFound
	}

	return launchVehicle, err
}

func (r *Repository) GetPublishedLaunchVehicles(
	minPayload,
	maxPayload int,
) ([]model.LaunchVehicle, error) {
	var launchVehicles []model.LaunchVehicle

	err := r.db.
		Preload("Likes").
		Where(
			"status = ? AND payload_kg BETWEEN ? AND ?",
			model.StatusPublished,
			minPayload,
			maxPayload,
		).
		Order("id ASC").
		Find(&launchVehicles).Error

	return launchVehicles, err
}

func (r *Repository) CreateDraft(
	creatorID uint,
	name string,
) (model.LaunchVehicle, error) {
	existing, err := r.GetDraftByCreator(creatorID)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, ErrDraftNotFound) {
		return model.LaunchVehicle{}, err
	}

	launchVehicle := model.LaunchVehicle{
		Name:      name,
		Status:    model.StatusDraft,
		ImageURL:  "",
		VideoURL:  "",
		CreatorID: creatorID,
	}

	if err := r.db.Create(&launchVehicle).Error; err != nil {
		return model.LaunchVehicle{}, err
	}

	return launchVehicle, nil
}

// PublishDraft — ORM.
func (r *Repository) PublishDraft(
	creatorID uint,
	id uint,
	shortDescription string,
	payloadKg int,
	seaLevelThrustKN int,
) error {
	formedAt := time.Now()

	result := r.db.
		Model(&model.LaunchVehicle{}).
		Where(
			"id = ? AND creator_id = ? AND status = ?",
			id,
			creatorID,
			model.StatusDraft,
		).
		Updates(map[string]any{
			"short_description":   shortDescription,
			"payload_kg":          payloadKg,
			"sea_level_thrust_kn": seaLevelThrustKN,
			"status":              model.StatusPublished,
			"formed_at":           &formedAt,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrDraftNotFound
	}

	return nil
}

func (r *Repository) DeleteLaunchVehicleRaw(
	ctx context.Context,
	id uint,
) error {
	result, err := r.sqlDB.ExecContext(
		ctx,
		`UPDATE launch_vehicles
		 SET status = $1
		 WHERE id = $2 AND status <> $1`,
		model.StatusDeleted,
		id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
