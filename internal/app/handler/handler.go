package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"orbitlab/internal/app/model"
	"orbitlab/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

const (
	currentUserID     uint = 1
	defaultMinPayload      = 0
	defaultMaxPayload      = 30000
	defaultImageURL        = "/static/defaults/default-launch-vehicle.png"
	defaultVideoURL        = "/static/defaults/default-launch-vehicle.mp4"
)

type Handler struct {
	Repository *repository.Repository
	httpClient *http.Client
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
		httpClient: &http.Client{Timeout: 700 * time.Millisecond},
	}
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	var id *uint

	idString := strings.TrimSpace(ctx.Query("id"))
	if idString != "" {
		parsed, err := strconv.ParseUint(idString, 10, 64)
		if err != nil || parsed == 0 {
			ctx.String(http.StatusBadRequest, "некорректный id")
			return
		}
		value := uint(parsed)
		id = &value
	}

	next := ctx.Query("next") == "true"

	launchVehicle, err := h.Repository.GetFeedLaunchVehicle(id, next)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.String(http.StatusNotFound, "launch vehicle не найден или удален")
		return
	}
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	launchVehicle = h.withMediaFallback(launchVehicle)

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"launchVehicle": launchVehicle,
		"likeCount":     len(launchVehicle.Likes),
	})
}

func (h *Handler) GetDraft(ctx *gin.Context) {
	launchVehicle, err := h.Repository.GetDraftByCreator(currentUserID)

	if errors.Is(err, repository.ErrDraftNotFound) {
		ctx.HTML(http.StatusOK, "add.html", gin.H{
			"hasDraft": false,
		})
		return
	}

	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	launchVehicle = h.withMediaFallback(launchVehicle)

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"hasDraft":      true,
		"launchVehicle": launchVehicle,
	})
}

func (h *Handler) GetLaunchVehicles(ctx *gin.Context) {
	minPayloadString := ctx.Query("payload_min")
	maxPayloadString := ctx.Query("payload_max")

	minPayload := defaultMinPayload
	maxPayload := defaultMaxPayload

	if minPayloadString != "" {
		parsed, err := strconv.Atoi(minPayloadString)
		if err != nil || parsed < 0 {
			ctx.String(http.StatusBadRequest, "payload_min должен быть неотрицательным целым числом")
			return
		}
		minPayload = parsed
	}

	if maxPayloadString != "" {
		parsed, err := strconv.Atoi(maxPayloadString)
		if err != nil || parsed < 0 {
			ctx.String(http.StatusBadRequest, "payload_max должен быть неотрицательным целым числом")
			return
		}
		maxPayload = parsed
	}

	if minPayload > maxPayload {
		ctx.String(http.StatusBadRequest, "payload_min не может быть больше payload_max")
		return
	}

	launchVehicles, err := h.Repository.GetPublishedLaunchVehicles(minPayload, maxPayload)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	likeCounts := make(map[uint]int)

	for i := range launchVehicles {
		launchVehicles[i] = h.withMediaFallback(launchVehicles[i])
		likeCounts[launchVehicles[i].ID] = len(launchVehicles[i].Likes)
	}

	ctx.HTML(http.StatusOK, "launch-vehicles.html", gin.H{
		"launchVehicles": launchVehicles,
		"likeCounts":     likeCounts,
		"minPayload":     minPayload,
		"maxPayload":     maxPayload,
	})
}

func (h *Handler) PostCreateDraft(ctx *gin.Context) {
	name := strings.TrimSpace(ctx.PostForm("name"))

	if name == "" {
		ctx.String(http.StatusBadRequest, "название обязательно")
		return
	}

	_, err := h.Repository.CreateDraft(currentUserID, name)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/add")
}

func (h *Handler) PostPublishDraft(ctx *gin.Context) {
	id, err := parseUint(ctx.PostForm("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, "некорректный id")
		return
	}

	payloadKg, err := strconv.Atoi(ctx.PostForm("payload_kg"))
	if err != nil || payloadKg < 0 {
		ctx.String(http.StatusBadRequest, "некорректная полезная нагрузка")
		return
	}

	seaLevelThrustKN, err := strconv.Atoi(ctx.PostForm("sea_level_thrust_kn"))
	if err != nil || seaLevelThrustKN < 0 {
		ctx.String(http.StatusBadRequest, "некорректная тяга на уровне моря")
		return
	}

	shortDescription := strings.TrimSpace(ctx.PostForm("short_description"))
	if shortDescription == "" {
		ctx.String(http.StatusBadRequest, "краткая информация обязательна")
		return
	}

	err = h.Repository.PublishDraft(
		currentUserID,
		id,
		shortDescription,
		payloadKg,
		seaLevelThrustKN,
	)
	if errors.Is(err, repository.ErrDraftNotFound) {
		ctx.String(http.StatusNotFound, "черновик не найден")
		return
	}
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/launch-vehicles")
}

func (h *Handler) PostDeleteLaunchVehicle(ctx *gin.Context) {
	id, err := parseUint(ctx.PostForm("id"))
	if err != nil {
		ctx.String(http.StatusBadRequest, "некорректный id")
		return
	}

	err = h.Repository.DeleteLaunchVehicleRaw(ctx.Request.Context(), id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		ctx.String(http.StatusNotFound, "launch vehicle не найден")
		return
	}
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/launch-vehicles")
}

func parseUint(raw string) (uint, error) {
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		return 0, errors.New("invalid uint")
	}
	return uint(value), nil
}

func (h *Handler) withMediaFallback(launchVehicle model.LaunchVehicle) model.LaunchVehicle {
	launchVehicle.ImageURL = h.resolveMedia(launchVehicle.ImageURL, defaultImageURL)
	launchVehicle.VideoURL = h.resolveMedia(launchVehicle.VideoURL, defaultVideoURL)
	return launchVehicle
}

func (h *Handler) resolveMedia(rawURL, fallback string) string {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return fallback
	}

	if strings.HasPrefix(rawURL, "/static/") {
		return rawURL
	}

	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		return fallback
	}

	request, err := http.NewRequest(http.MethodHead, rawURL, nil)
	if err != nil {
		return fallback
	}

	response, err := h.httpClient.Do(request)
	if err != nil {
		return fallback
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusBadRequest {
		return fallback
	}

	return rawURL
}
