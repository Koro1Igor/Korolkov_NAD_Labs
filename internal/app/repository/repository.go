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
		getEnv("PGPORT", "5433"),
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к PostgreSQL: %w", err)
	}

	// Для уже созданной БД сначала добавляем password без NOT NULL,
	// заполняем существующие строки и только затем включаем ограничение.
	if db.Migrator().HasTable(&model.User{}) {
		if !db.Migrator().HasColumn(&model.User{}, "Password") {
			if err := db.Exec(`ALTER TABLE users ADD COLUMN password varchar(255)`).Error; err != nil {
				return nil, fmt.Errorf("ошибка добавления users.password: %w", err)
			}
		}
		if err := db.Exec(`UPDATE users SET password = 'change-me' WHERE password IS NULL`).Error; err != nil {
			return nil, fmt.Errorf("ошибка заполнения users.password: %w", err)
		}
		if err := db.Exec(`ALTER TABLE users ALTER COLUMN password SET NOT NULL`).Error; err != nil {
			return nil, fmt.Errorf("ошибка ограничения users.password: %w", err)
		}
	}

	// created_at нужен только у launch_vehicles. Для существующей БД
	// сначала добавляем nullable-колонку, заполняем строки и ставим NOT NULL.
	if db.Migrator().HasTable(&model.LaunchVehicle{}) {
		if !db.Migrator().HasColumn(&model.LaunchVehicle{}, "CreatedAt") {
			if err := db.Exec(`ALTER TABLE launch_vehicles ADD COLUMN created_at timestamptz`).Error; err != nil {
				return nil, fmt.Errorf("ошибка добавления launch_vehicles.created_at: %w", err)
			}
		}
		if err := db.Exec(`UPDATE launch_vehicles SET created_at = COALESCE(formed_at, NOW()) WHERE created_at IS NULL`).Error; err != nil {
			return nil, fmt.Errorf("ошибка заполнения launch_vehicles.created_at: %w", err)
		}
		if err := db.Exec(`ALTER TABLE launch_vehicles ALTER COLUMN created_at SET NOT NULL`).Error; err != nil {
			return nil, fmt.Errorf("ошибка ограничения launch_vehicles.created_at: %w", err)
		}
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.LaunchVehicle{},
		&model.Like{},
	); err != nil {
		return nil, fmt.Errorf("ошибка миграции: %w", err)
	}

	// Эти поля отсутствуют в согласованной ER-модели. Необязательные поля
	// launch_vehicles должны принимать NULL, в том числе в черновике.
	statements := []string{
		`ALTER TABLE users DROP COLUMN IF EXISTS created_at`,
		`ALTER TABLE likes DROP COLUMN IF EXISTS created_at`,
		`ALTER TABLE launch_vehicles ALTER COLUMN short_description DROP NOT NULL`,
		`ALTER TABLE launch_vehicles ALTER COLUMN image_url DROP NOT NULL`,
		`ALTER TABLE launch_vehicles ALTER COLUMN video_url DROP NOT NULL`,
		`ALTER TABLE launch_vehicles ALTER COLUMN payload_kg DROP NOT NULL`,
		`ALTER TABLE launch_vehicles ALTER COLUMN payload_kg DROP DEFAULT`,
		`ALTER TABLE launch_vehicles ALTER COLUMN sea_level_thrust_kn DROP NOT NULL`,
		`ALTER TABLE launch_vehicles ALTER COLUMN sea_level_thrust_kn DROP DEFAULT`,
		`ALTER TABLE launch_vehicles ALTER COLUMN formed_at DROP NOT NULL`,
		`ALTER TABLE launch_vehicles DROP CONSTRAINT IF EXISTS fk_launch_vehicles_creator`,
		`ALTER TABLE launch_vehicles ADD CONSTRAINT fk_launch_vehicles_creator FOREIGN KEY (creator_id) REFERENCES users(id) ON DELETE RESTRICT`,
		`ALTER TABLE likes DROP CONSTRAINT IF EXISTS fk_likes_user`,
		`ALTER TABLE likes ADD CONSTRAINT fk_likes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT`,
		`ALTER TABLE likes DROP CONSTRAINT IF EXISTS fk_launch_vehicles_likes`,
		`ALTER TABLE likes ADD CONSTRAINT fk_launch_vehicles_likes FOREIGN KEY (launch_vehicle_id) REFERENCES launch_vehicles(id) ON DELETE RESTRICT`,
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		for _, statement := range statements {
			if err := tx.Exec(statement).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("ошибка приведения схемы БД: %w", err)
	}

	// У одного пользователя не может быть более одного launch vehicle
	// в статусе draft.
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

// GetFeedLaunchVehicle получает для ленты сразу одну строку из БД.
// Удаленные launch vehicles никогда не возвращаются.
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

	// Если после текущей карточки опубликованных записей нет,
	// возвращаем первую опубликованную карточку.
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

// CreateDraft — ORM.
func (r *Repository) CreateDraft(
	creatorID uint,
	name string,
) (model.LaunchVehicle, error) {
	// Если у пользователя уже есть draft, второй не создаем.
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

// DeleteLaunchVehicleRaw — SQL UPDATE без ORM.
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
