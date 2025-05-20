package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/config"
	"github.com/morzisorn/metrics/internal/server/services/health"
)

type HealthController struct {
	service *health.HealthService
}

// NewHealthControlle receives Health service and creates new Health controller, returns pointer.
func NewHealthController(service *health.HealthService) *HealthController {
	return &HealthController{service: service}
}

// PingDB checks if DB is alive.
func (hc *HealthController) PingDB(c *gin.Context) {
	service := config.GetService("server")

	if service.Config.DBConnStr == "" {
		c.Status(http.StatusInternalServerError)
		return
	}

	if err := health.PingDB(); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}
