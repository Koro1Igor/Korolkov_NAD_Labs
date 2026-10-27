package repository

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"orbitlab/internal/app/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

type ObjectStorage struct {
	endpoint  string
	accessKey string
	secretKey string
	bucket    string
	region    string
	client    *http.Client
}

func NewRepository() (*Repository, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Europe/Moscow",
		env("PGHOST", "localhost"),
		env("PGUSER", "orbit"),
		env("PGPASSWORD", "orbit12345"),
		env("PGDATABASE", "orbitlab"),
		env("PGPORT", "5433"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres connection: %w", err)
	}
	if db.Migrator().HasTable("launch_vehicles") {
		if db.Migrator().HasColumn("launch_vehicles", "image_url") && !db.Migrator().HasColumn("launch_vehicles", "image_name") {
			if err := db.Migrator().RenameColumn("launch_vehicles", "image_url", "image_name"); err != nil {
				return nil, err
			}
		}
		if db.Migrator().HasColumn("launch_vehicles", "video_url") && !db.Migrator().HasColumn("launch_vehicles", "video_name") {
			if err := db.Migrator().RenameColumn("launch_vehicles", "video_url", "video_name"); err != nil {
				return nil, err
			}
		}
	}
	if err := db.AutoMigrate(&model.User{}, &model.LaunchVehicle{}, &model.Like{}); err != nil {
		return nil, fmt.Errorf("migration: %w", err)
	}
	statements := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_one_draft_per_creator ON launch_vehicles (creator_id) WHERE status = 'draft'`,
		`UPDATE launch_vehicles SET image_name = regexp_replace(image_name, '^https?://[^/]+/service-media/', '') WHERE image_name LIKE 'http%'`,
		`UPDATE launch_vehicles SET video_name = regexp_replace(video_name, '^https?://[^/]+/service-media/', '') WHERE video_name LIKE 'http%'`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return nil, fmt.Errorf("schema setup: %w", err)
		}
	}
	return &Repository{db: db}, nil
}

func NewObjectStorage() *ObjectStorage {
	return &ObjectStorage{
		endpoint:  env("MINIO_ENDPOINT", "http://localhost:9000"),
		accessKey: env("MINIO_ACCESS_KEY", "minioadmin"),
		secretKey: env("MINIO_SECRET_KEY", "minioadmin123"),
		bucket:    env("MINIO_BUCKET", "service-media"),
		region:    env("MINIO_REGION", "us-east-1"),
		client:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (r *Repository) ListPublished(ctx context.Context, minPayload, maxPayload int) ([]model.LaunchVehicle, error) {
	var items []model.LaunchVehicle
	err := r.db.WithContext(ctx).Preload("Likes").Where("status = ? AND payload_kg BETWEEN ? AND ?", model.StatusPublished, minPayload, maxPayload).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repository) Feed(ctx context.Context, id *uint, next bool) (model.LaunchVehicle, error) {
	var item model.LaunchVehicle
	query := r.db.WithContext(ctx).Preload("Likes").Where("status = ?", model.StatusPublished)
	if id == nil {
		return item, normalizeNotFound(query.Order("id ASC").First(&item).Error)
	}
	if !next {
		return item, normalizeNotFound(query.Where("id = ?", *id).First(&item).Error)
	}
	err := query.Where("id > ?", *id).Order("id ASC").First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = r.db.WithContext(ctx).Preload("Likes").Where("status = ?", model.StatusPublished).Order("id ASC").First(&item).Error
	}
	return item, normalizeNotFound(err)
}

func (r *Repository) DraftByCreator(ctx context.Context, creatorID uint) (model.LaunchVehicle, error) {
	var item model.LaunchVehicle
	err := r.db.WithContext(ctx).Preload("Likes").Where("creator_id = ? AND status = ?", creatorID, model.StatusDraft).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, model.ErrDraftNotFound
	}
	return item, err
}

func (r *Repository) CreateDraft(ctx context.Context, creatorID uint, name, imageName, videoName string) (model.LaunchVehicle, error) {
	if _, err := r.DraftByCreator(ctx, creatorID); err == nil {
		return model.LaunchVehicle{}, model.ErrDraftExists
	} else if !errors.Is(err, model.ErrDraftNotFound) {
		return model.LaunchVehicle{}, err
	}
	item := model.LaunchVehicle{Name: name, CreatorID: creatorID, Status: model.StatusDraft, ImageName: &imageName, VideoName: &videoName}
	if err := r.db.WithContext(ctx).Create(&item).Error; err != nil {
		return model.LaunchVehicle{}, err
	}
	return item, nil
}

func (r *Repository) PublishDraft(ctx context.Context, creatorID uint, input model.PublishLaunchVehicleInput) (model.LaunchVehicle, error) {
	var item model.LaunchVehicle
	err := r.db.WithContext(ctx).Where("creator_id = ? AND status = ?", creatorID, model.StatusDraft).First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return item, model.ErrDraftNotFound
	}
	if err != nil {
		return item, err
	}
	result := r.db.WithContext(ctx).Model(&model.LaunchVehicle{}).Where("id = ? AND creator_id = ? AND status = ?", item.ID, creatorID, model.StatusDraft).Updates(map[string]any{
		"short_description":   input.ShortDescription,
		"payload_kg":          input.PayloadKg,
		"sea_level_thrust_kn": input.SeaLevelThrustKN,
		"status":              model.StatusPublished,
		"formed_at":           time.Now(),
	})
	if result.Error != nil {
		return model.LaunchVehicle{}, result.Error
	}
	if result.RowsAffected == 0 {
		return model.LaunchVehicle{}, model.ErrDraftNotFound
	}
	return r.ByID(ctx, item.ID)
}

func (r *Repository) SoftDeleteOwn(ctx context.Context, creatorID, id uint) error {
	result := r.db.WithContext(ctx).Model(&model.LaunchVehicle{}).Where("id = ? AND creator_id = ? AND status IN ?", id, creatorID, []string{model.StatusDraft, model.StatusPublished}).Updates(map[string]any{"status": model.StatusDeleted, "completed_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *Repository) SetLike(ctx context.Context, userID, launchVehicleID uint, value int) (model.LikeResult, error) {
	var vehicle model.LaunchVehicle
	if err := r.db.WithContext(ctx).Where("id = ? AND status = ?", launchVehicleID, model.StatusPublished).First(&vehicle).Error; err != nil {
		return model.LikeResult{}, normalizeNotFound(err)
	}
	if value == 1 {
		like := model.Like{UserID: userID, LaunchVehicleID: launchVehicleID}
		if err := r.db.WithContext(ctx).Where("user_id = ? AND launch_vehicle_id = ?", userID, launchVehicleID).FirstOrCreate(&like).Error; err != nil {
			return model.LikeResult{}, err
		}
	} else if err := r.db.WithContext(ctx).Where("user_id = ? AND launch_vehicle_id = ?", userID, launchVehicleID).Delete(&model.Like{}).Error; err != nil {
		return model.LikeResult{}, err
	}
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Like{}).Where("launch_vehicle_id = ?", launchVehicleID).Count(&count).Error; err != nil {
		return model.LikeResult{}, err
	}
	return model.LikeResult{LaunchVehicleID: launchVehicleID, Liked: value, LikeCount: count}, nil
}

func (r *Repository) ByID(ctx context.Context, id uint) (model.LaunchVehicle, error) {
	var item model.LaunchVehicle
	err := r.db.WithContext(ctx).Preload("Likes").Where("id = ? AND status <> ?", id, model.StatusDeleted).First(&item).Error
	return item, normalizeNotFound(err)
}

func (r *Repository) CreateUser(ctx context.Context, input model.RegisterUserInput) (model.User, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", input.Email).Count(&count).Error; err != nil {
		return model.User{}, err
	}
	if count > 0 {
		return model.User{}, model.ErrEmailExists
	}
	user := model.User{Name: input.Name, Email: input.Email, Password: input.Password}
	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}

func (r *Repository) Authenticate(ctx context.Context, email, password string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ? AND password = ?", email, password).First(&user).Error
	return user, normalizeNotFound(err)
}

func (s *ObjectStorage) Upload(ctx context.Context, objectName, contentType string, reader io.Reader, size int64) error {
	if size < 0 {
		return model.ErrInvalidInput
	}
	data, err := io.ReadAll(io.LimitReader(reader, size+1))
	if err != nil {
		return err
	}
	if int64(len(data)) != size {
		return fmt.Errorf("unexpected file size")
	}
	return s.do(ctx, http.MethodPut, objectName, contentType, data)
}

func (s *ObjectStorage) Delete(ctx context.Context, objectName string) error {
	return s.do(ctx, http.MethodDelete, objectName, "", nil)
}

func (s *ObjectStorage) Endpoint() string { return s.endpoint }

func (s *ObjectStorage) Bucket() string { return s.bucket }

func (s *ObjectStorage) do(ctx context.Context, method, objectName, contentType string, payload []byte) error {
	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	date := now.Format("20060102")
	payloadHash := sha256Hex(payload)
	base, err := url.Parse(strings.TrimRight(s.endpoint, "/"))
	if err != nil {
		return err
	}
	canonicalURI := "/" + url.PathEscape(s.bucket) + "/" + escapeObjectPath(objectName)
	requestURL := *base
	requestURL.Path = path.Join(base.Path, s.bucket, objectName)
	headerLines := []string{}
	signedHeaders := []string{}
	if contentType != "" {
		headerLines = append(headerLines, "content-type:"+strings.TrimSpace(contentType))
		signedHeaders = append(signedHeaders, "content-type")
	}
	headerLines = append(headerLines, "host:"+base.Host, "x-amz-content-sha256:"+payloadHash, "x-amz-date:"+amzDate)
	signedHeaders = append(signedHeaders, "host", "x-amz-content-sha256", "x-amz-date")
	canonicalHeaders := strings.Join(headerLines, "\n") + "\n"
	signedHeaderString := strings.Join(signedHeaders, ";")
	canonicalRequest := strings.Join([]string{method, canonicalURI, "", canonicalHeaders, signedHeaderString, payloadHash}, "\n")
	scope := date + "/" + s.region + "/s3/aws4_request"
	stringToSign := strings.Join([]string{"AWS4-HMAC-SHA256", amzDate, scope, sha256Hex([]byte(canonicalRequest))}, "\n")
	kDate := hmacSHA256([]byte("AWS4"+s.secretKey), date)
	kRegion := hmacSHA256(kDate, s.region)
	kService := hmacSHA256(kRegion, "s3")
	kSigning := hmacSHA256(kService, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(kSigning, stringToSign))
	authorization := "AWS4-HMAC-SHA256 Credential=" + s.accessKey + "/" + scope + ", SignedHeaders=" + signedHeaderString + ", Signature=" + signature
	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("Authorization", authorization)
	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return fmt.Errorf("minio %s: %s %s", method, resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

func normalizeNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrNotFound
	}
	return err
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func escapeObjectPath(value string) string {
	parts := strings.Split(strings.TrimLeft(value, "/"), "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
