package api

import (
	"context"
	"log"
	"net/http"

	"orbitlab/internal/app/auth"
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
	authService, err := auth.NewService(context.Background())
	if err != nil {
		logrus.Fatal(err)
	}
	defer authService.Close()
	h := handler.NewHandler(repo, storage, authService)

	r := gin.Default()
	r.Use(cors())
	r.LoadHTMLGlob("templates/*.html")
	r.Static("/resources", "./resources")
	r.GET("/", func(ctx *gin.Context) { ctx.HTML(200, "feed.html", nil) })
	r.GET("/launch-vehicles", func(ctx *gin.Context) { ctx.HTML(200, "launch-vehicles.html", nil) })
	r.GET("/add", func(ctx *gin.Context) { ctx.HTML(200, "add.html", nil) })
	r.GET("/swagger/doc.yaml", swaggerJSON)
	r.GET("/swagger/index.html", swaggerUI)

	api := r.Group("/api")
	api.GET("/launch-vehicles", h.OptionalAuth, h.GetLaunchVehicles)
	api.GET("/feed", h.OptionalAuth, h.GetFeed)
	api.POST("/users", h.RegisterUser)
	api.POST("/users/login", h.Login)

	protected := api.Group("")
	protected.Use(h.RequireAuth)
	protected.GET("/launch-vehicles/draft", h.GetDraft)
	protected.POST("/launch-vehicles", h.CreateLaunchVehicle)
	protected.PUT("/launch-vehicles/draft", h.PublishDraft)
	protected.DELETE("/launch-vehicles/:id", h.DeleteLaunchVehicle)
	protected.POST("/launch-vehicles/:id/like", h.SetLike)
	protected.POST("/users/logout", h.Logout)

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

func swaggerUI(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.String(http.StatusOK, `<!doctype html><html><head><meta charset="utf-8"><title>OrbitLab Swagger</title><link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css"></head><body><div id="swagger-ui"></div><script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script><script>SwaggerUIBundle({url:'/swagger/doc.yaml',dom_id:'#swagger-ui',persistAuthorization:true})</script></body></html>`)
}
