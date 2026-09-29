package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"orbitlab/internal/app/repository"
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{Repository: r}
}

func likeCount(service repository.Service) int {
	return len(service.Likes)
}

func (h *Handler) GetFeed(ctx *gin.Context) {
	idString := ctx.Query("id")
	next := ctx.Query("next") == "true"

	var service repository.Service
	var err error

	if idString == "" {
		service, err = h.Repository.GetFirstPublishedService()
	} else {
		id, parseErr := strconv.Atoi(idString)
		if parseErr != nil {
			ctx.String(http.StatusBadRequest, "некорректный id")
			return
		}
		if next {
			service, err = h.Repository.GetNextPublishedService(id)
		} else {
			service, err = h.Repository.GetServiceByID(id)
		}
	}

	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.HTML(http.StatusOK, "feed.html", gin.H{
		"service":   service,
		"likeCount": likeCount(service),
	})
}

func (h *Handler) GetDraft(ctx *gin.Context) {
	service, err := h.Repository.GetDraftService()
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusNotFound, err.Error())
		return
	}

	ctx.HTML(http.StatusOK, "add.html", gin.H{
		"service": service,
	})
}

func (h *Handler) GetServices(ctx *gin.Context) {
	filterString := ctx.Query("min_payload")
	minPayload := 0

	if filterString != "" {
		parsed, err := strconv.Atoi(filterString)
		if err != nil || parsed < 0 {
			ctx.String(http.StatusBadRequest, "min_payload должен быть неотрицательным целым числом")
			return
		}
		minPayload = parsed
	}

	services, err := h.Repository.GetFilteredServices(minPayload)
	if err != nil {
		logrus.Error(err)
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}

	likeCounts := make(map[int]int)
	for _, service := range services {
		likeCounts[service.ID] = likeCount(service)
	}

	ctx.HTML(http.StatusOK, "services.html", gin.H{
		"services":   services,
		"likeCounts": likeCounts,
		"minPayload": filterString,
	})
}
