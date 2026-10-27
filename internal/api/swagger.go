package api

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func swaggerJSON(ctx *gin.Context) {
	data, err := os.ReadFile("docs/openapi.yaml")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Data(http.StatusOK, "application/yaml; charset=utf-8", data)
}
