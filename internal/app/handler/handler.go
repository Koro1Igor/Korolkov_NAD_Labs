package handler

import (
	"net/http"
	"strconv"

	"orbitlab/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	defaultMinPayload = 0
	defaultMaxPayload = 30000
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

func likeCount(launchVehicle repository.LaunchVehicle) int {
	return len(launchVehicle.Likes)
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	idString := ctx.Query("id")
	next := ctx.Query("next") == "true"

	var launchVehicle repository.LaunchVehicle
	var err error

	if idString == "" {
		launchVehicle, err = h.Repository.GetFirstPublishedLaunchVehicle()
	} else {
		id, parseErr := strconv.Atoi(idString)
		if parseErr != nil {
			ctx.String(http.StatusBadRequest, "некорректный id")
			return
		}
		if next {
			launchVehicle, err = h.Repository.GetNextPublishedLaunchVehicle(id)
		} else {
			launchVehicle, err = h.Repository.GetLaunchVehicleByID(id)
		}
	}

	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"launchVehicle": launchVehicle,
		"likeCount":     likeCount(launchVehicle),
	})
}

func (h *Handler) GetDraft(ctx *gin.Context) {
	launchVehicle, err := h.Repository.GetDraftLaunchVehicle()
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
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

	launchVehicles, err := h.Repository.GetFilteredLaunchVehicles(minPayload, maxPayload)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	likeCounts := make(map[int]int)
	for _, launchVehicle := range launchVehicles {
		likeCounts[launchVehicle.ID] = likeCount(launchVehicle)
	}

	ctx.HTML(http.StatusOK, "launch-vehicles.html", gin.H{
		"launchVehicles": launchVehicles,
		"likeCounts":     likeCounts,
		"minPayload":     minPayload,
		"maxPayload":     maxPayload,
	})
}
