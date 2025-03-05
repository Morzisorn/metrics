package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/storage/database"
)

func PingDB(c *gin.Context) {
	service := config.GetService("server")

	if service.Config.DBConnStr == "" {
		c.Status(http.StatusInternalServerError)
		return
	}

	db := database.GetDB()
	if err := database.PingDB(db); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
