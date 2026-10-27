package api

import (
	"log"

	"orbitlab/internal/app/handler"
	"orbitlab/internal/app/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func StartServer() {
	repo, err := repository.NewRepository()
	if err != nil {
		logrus.Fatal(err)
	}
	storage := repository.NewObjectStorage()
	h := handler.NewHandler(repo, storage)

	r := gin.Default()
	r.Use(cors())
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/resources", "./resources")
	r.GET("/", func(ctx *gin.Context) { ctx.HTML(200, "feed.html", nil) })
	r.GET("/launch-vehicles", func(ctx *gin.Context) { ctx.HTML(200, "launch-vehicles.html", nil) })
	r.GET("/add", func(ctx *gin.Context) { ctx.HTML(200, "add.html", nil) })

	api := r.Group("/api")
	api.GET("/launch-vehicles", h.GetLaunchVehicles)
	api.GET("/launch-vehicles/draft", h.GetDraft)
	api.POST("/launch-vehicles", h.CreateLaunchVehicle)
	api.PUT("/launch-vehicles/draft", h.PublishDraft)
	api.DELETE("/launch-vehicles/:id", h.DeleteLaunchVehicle)
	api.POST("/launch-vehicles/:id/like", h.SetLike)
	api.GET("/feed", h.GetFeed)
	api.POST("/users", h.RegisterUser)
	api.POST("/users/login", h.Login)
	api.POST("/users/logout", h.Logout)

	log.Println("Listening on :8080")
	if err := r.Run(":8080"); err != nil {
		logrus.Fatal(err)
	}
}

func cors() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Next()
	}
}
