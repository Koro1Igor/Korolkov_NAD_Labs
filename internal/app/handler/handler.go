package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"orbitlab/internal/app/auth"
	"orbitlab/internal/app/model"
	"orbitlab/internal/app/repository"
	"orbitlab/internal/app/serializer"

	"github.com/gin-gonic/gin"
)

const (
	defaultMinPayload = 0
	defaultMaxPayload = 30000
	maxImageSize      = 10 * 1024 * 1024
	maxVideoSize      = 60 * 1024 * 1024
)

type Handler struct {
	repository *repository.Repository
	storage    *repository.ObjectStorage
	auth       *auth.Service
}

func NewHandler(repo *repository.Repository, storage *repository.ObjectStorage, authService *auth.Service) *Handler {
	return &Handler{repository: repo, storage: storage, auth: authService}
}

func (h *Handler) OptionalAuth(ctx *gin.Context) {
	tokenString := bearerToken(ctx)
	if tokenString == "" {
		ctx.Set("user_id", uint(0))
		ctx.Next()
		return
	}
	claims, err := h.auth.Parse(ctx.Request.Context(), tokenString)
	if err != nil {
		ctx.Set("user_id", uint(0))
		ctx.Next()
		return
	}
	ctx.Set("user_id", claims.UserID)
	ctx.Set("jwt", tokenString)
	ctx.Next()
}

func (h *Handler) RequireAuth(ctx *gin.Context) {
	tokenString := bearerToken(ctx)
	if tokenString == "" {
		writeError(ctx, http.StatusUnauthorized, "authorization required")
		ctx.Abort()
		return
	}
	claims, err := h.auth.Parse(ctx.Request.Context(), tokenString)
	if err != nil {
		writeError(ctx, http.StatusUnauthorized, "invalid or expired token")
		ctx.Abort()
		return
	}
	ctx.Set("user_id", claims.UserID)
	ctx.Set("jwt", tokenString)
	ctx.Next()
}

func (h *Handler) GetLaunchVehicles(ctx *gin.Context) {
	minPayload, err := queryInt(ctx, "payload_min", defaultMinPayload)
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid payload_min")
		return
	}
	maxPayload, err := queryInt(ctx, "payload_max", defaultMaxPayload)
	if err != nil || minPayload > maxPayload {
		writeError(ctx, http.StatusBadRequest, "invalid payload_max")
		return
	}
	userID := currentUserID(ctx)
	items, err := h.repository.ListPublished(ctx.Request.Context(), minPayload, maxPayload)
	if err != nil {
		handleError(ctx, err)
		return
	}
	result := make([]serializer.LaunchVehicleListItem, 0, len(items))
	for _, item := range items {
		result = append(result, serializer.ListItem(item, userID, h.storage.Endpoint(), h.storage.Bucket()))
	}
	ctx.JSON(http.StatusOK, gin.H{"items": result})
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	var id *uint
	if raw := strings.TrimSpace(ctx.Query("id")); raw != "" {
		value, err := parseUint(raw)
		if err != nil {
			writeError(ctx, http.StatusBadRequest, "invalid id")
			return
		}
		id = &value
	}
	next := false
	if raw := strings.TrimSpace(ctx.Query("next")); raw != "" {
		value, err := strconv.ParseBool(raw)
		if err != nil {
			writeError(ctx, http.StatusBadRequest, "invalid next")
			return
		}
		next = value
	}
	if next && id == nil {
		writeError(ctx, http.StatusBadRequest, "id is required when next=true")
		return
	}
	item, err := h.repository.Feed(ctx.Request.Context(), id, next)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.FeedItem(item, currentUserID(ctx), h.storage.Endpoint(), h.storage.Bucket()))
}

func (h *Handler) GetDraft(ctx *gin.Context) {
	item, err := h.repository.DraftByCreator(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.Detail(item, h.storage.Endpoint(), h.storage.Bucket()))
}

func (h *Handler) CreateLaunchVehicle(ctx *gin.Context) {
	userID := currentUserID(ctx)
	if err := ctx.Request.ParseMultipartForm(maxVideoSize + maxImageSize + 1<<20); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid multipart form")
		return
	}
	for field := range ctx.Request.MultipartForm.Value {
		if field != "name" {
			writeError(ctx, http.StatusBadRequest, "unknown or system field: "+field)
			return
		}
	}
	for field := range ctx.Request.MultipartForm.File {
		if field != "image" && field != "video" {
			writeError(ctx, http.StatusBadRequest, "unknown file field: "+field)
			return
		}
	}
	name := strings.TrimSpace(ctx.PostForm("name"))
	if name == "" || len([]rune(name)) > 120 {
		writeError(ctx, http.StatusBadRequest, "invalid name")
		return
	}
	imageHeader, err := ctx.FormFile("image")
	if err != nil || imageHeader.Size <= 0 || imageHeader.Size > maxImageSize {
		writeError(ctx, http.StatusBadRequest, "invalid image file")
		return
	}
	videoHeader, err := ctx.FormFile("video")
	if err != nil || videoHeader.Size <= 0 || videoHeader.Size > maxVideoSize {
		writeError(ctx, http.StatusBadRequest, "invalid video file")
		return
	}
	imageExt := normalizedExtension(imageHeader.Filename, true)
	videoExt := normalizedExtension(videoHeader.Filename, false)
	if imageExt == "" || videoExt == "" {
		writeError(ctx, http.StatusBadRequest, "unsupported media extension")
		return
	}
	if _, err := h.repository.DraftByCreator(ctx.Request.Context(), userID); err == nil {
		handleError(ctx, model.ErrDraftExists)
		return
	} else if !errors.Is(err, model.ErrDraftNotFound) {
		handleError(ctx, err)
		return
	}
	image, err := imageHeader.Open()
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "cannot open image")
		return
	}
	defer image.Close()
	video, err := videoHeader.Open()
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "cannot open video")
		return
	}
	defer video.Close()
	imageName := generateObjectName("images", "image", imageExt)
	videoName := generateObjectName("videos", "video", videoExt)
	requestContext := ctx.Request.Context()
	if err := h.storage.Upload(requestContext, imageName, contentType(imageHeader.Header.Get("Content-Type"), imageExt), image, imageHeader.Size); err != nil {
		handleError(ctx, err)
		return
	}
	if err := h.storage.Upload(requestContext, videoName, contentType(videoHeader.Header.Get("Content-Type"), videoExt), video, videoHeader.Size); err != nil {
		_ = h.storage.Delete(requestContext, imageName)
		handleError(ctx, err)
		return
	}
	item, err := h.repository.CreateDraft(requestContext, userID, name, imageName, videoName)
	if err != nil {
		_ = h.storage.Delete(requestContext, imageName)
		_ = h.storage.Delete(requestContext, videoName)
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, serializer.Detail(item, h.storage.Endpoint(), h.storage.Bucket()))
}

func (h *Handler) PublishDraft(ctx *gin.Context) {
	var input model.PublishLaunchVehicleInput
	if err := decodeJSON(ctx, &input); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid json")
		return
	}
	input.ShortDescription = strings.TrimSpace(input.ShortDescription)
	if input.ShortDescription == "" || len([]rune(input.ShortDescription)) > 500 || input.PayloadKg < 0 || input.SeaLevelThrustKN < 0 {
		writeError(ctx, http.StatusBadRequest, "invalid input")
		return
	}
	item, err := h.repository.PublishDraft(ctx.Request.Context(), currentUserID(ctx), input)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.Detail(item, h.storage.Endpoint(), h.storage.Bucket()))
}

func (h *Handler) DeleteLaunchVehicle(ctx *gin.Context) {
	id, err := parseUint(ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.repository.SoftDeleteOwn(ctx.Request.Context(), currentUserID(ctx), id); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (h *Handler) SetLike(ctx *gin.Context) {
	id, err := parseUint(ctx.Param("id"))
	if err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid id")
		return
	}
	var input model.LikeInput
	if err := decodeJSON(ctx, &input); err != nil || input.Value < 0 || input.Value > 1 {
		writeError(ctx, http.StatusBadRequest, "value must be 0 or 1")
		return
	}
	result, err := h.repository.SetLike(ctx.Request.Context(), currentUserID(ctx), id, input.Value)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.LikeResponse{ID: result.LaunchVehicleID, Liked: result.Liked, LikeCount: result.LikeCount})
}

func (h *Handler) RegisterUser(ctx *gin.Context) {
	var input model.RegisterUserInput
	if err := decodeJSON(ctx, &input); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid json")
		return
	}
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Password = strings.TrimSpace(input.Password)
	if input.Name == "" || len([]rune(input.Name)) > 100 || len(input.Password) < 4 || len(input.Password) > 255 {
		writeError(ctx, http.StatusBadRequest, "invalid input")
		return
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid email")
		return
	}
	user, err := h.repository.CreateUser(ctx.Request.Context(), input)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, serializer.User(user))
}

func (h *Handler) Login(ctx *gin.Context) {
	var input model.LoginInput
	if err := decodeJSON(ctx, &input); err != nil {
		writeError(ctx, http.StatusBadRequest, "invalid json")
		return
	}
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	input.Password = strings.TrimSpace(input.Password)
	if input.Email == "" || input.Password == "" {
		writeError(ctx, http.StatusBadRequest, "invalid input")
		return
	}
	user, err := h.repository.Authenticate(ctx.Request.Context(), input.Email, input.Password)
	if err != nil {
		handleError(ctx, err)
		return
	}
	token, expiresAt, err := h.auth.Issue(user.ID)
	if err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, serializer.LoginResponse{AccessToken: token, TokenType: "Bearer", ExpiresAt: expiresAt, User: serializer.User(user)})
}

func (h *Handler) Logout(ctx *gin.Context) {
	tokenString, _ := ctx.Get("jwt")
	if tokenString == nil {
		writeError(ctx, http.StatusUnauthorized, "authorization required")
		return
	}
	if err := h.auth.Blacklist(ctx.Request.Context(), tokenString.(string)); err != nil {
		handleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"authenticated": 0})
}

func currentUserID(ctx *gin.Context) uint {
	value, ok := ctx.Get("user_id")
	if !ok {
		return 0
	}
	id, _ := value.(uint)
	return id
}
func bearerToken(ctx *gin.Context) string {
	value := strings.TrimSpace(ctx.GetHeader("Authorization"))
	const prefix = "Bearer "
	if !strings.HasPrefix(value, prefix) {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(value, prefix))
}
func queryInt(ctx *gin.Context, name string, fallback int) (int, error) {
	raw := strings.TrimSpace(ctx.Query(name))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0, model.ErrInvalidInput
	}
	return value, nil
}
func parseUint(raw string) (uint, error) {
	value, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || value == 0 {
		return 0, model.ErrInvalidInput
	}
	return uint(value), nil
}
func decodeJSON(ctx *gin.Context, target any) error {
	decoder := json.NewDecoder(io.LimitReader(ctx.Request.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if decoder.Decode(&struct{}{}) != io.EOF {
		return model.ErrInvalidInput
	}
	return nil
}
func handleError(ctx *gin.Context, err error) {
	switch {
	case errors.Is(err, model.ErrInvalidInput):
		writeError(ctx, http.StatusBadRequest, "invalid input")
	case errors.Is(err, model.ErrUnauthorized):
		writeError(ctx, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, model.ErrDraftExists):
		writeError(ctx, http.StatusConflict, "draft already exists")
	case errors.Is(err, model.ErrDraftNotFound):
		writeError(ctx, http.StatusNotFound, "draft not found")
	case errors.Is(err, model.ErrEmailExists):
		writeError(ctx, http.StatusConflict, "email already exists")
	case errors.Is(err, model.ErrNotFound):
		writeError(ctx, http.StatusNotFound, "not found")
	default:
		writeError(ctx, http.StatusInternalServerError, err.Error())
	}
}
func writeError(ctx *gin.Context, status int, message string) {
	ctx.JSON(status, gin.H{"error": message})
}
func generateObjectName(folder, prefix, extension string) string {
	buffer := make([]byte, 6)
	_, _ = rand.Read(buffer)
	return fmt.Sprintf("%s/%s-%d-%s%s", folder, prefix, time.Now().UnixNano(), hex.EncodeToString(buffer), extension)
}
func normalizedExtension(name string, image bool) string {
	extension := strings.ToLower(filepath.Ext(name))
	if image {
		switch extension {
		case ".jpg", ".jpeg", ".png", ".webp":
			return extension
		}
		return ""
	}
	switch extension {
	case ".mp4", ".mov", ".webm":
		return extension
	}
	return ""
}
func contentType(value, extension string) string {
	value = strings.TrimSpace(value)
	if value != "" && value != "application/octet-stream" {
		return value
	}
	if detected := mime.TypeByExtension(extension); detected != "" {
		return detected
	}
	return "application/octet-stream"
}
