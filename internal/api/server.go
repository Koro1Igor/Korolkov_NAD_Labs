package api

import (
	"log"

	"orbitlab/internal/app/handler"
	"orbitlab/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	log.Println("Starting server")

	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Fatal("ошибка инициализации репозитория: ", err)
	}

	h := handler.NewHandler(repo)

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.Static("/static", "./resources")

	r.GET("/feed", h.GetFeed)
	r.GET("/add", h.GetDraft)
	r.GET("/launch-vehicles", h.GetLaunchVehicles)

	if err := r.Run(":8080"); err != nil {
		logrus.Fatal(err)
	}
}
