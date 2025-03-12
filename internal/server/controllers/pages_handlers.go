package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/internal/server/services/pages"
)

type PagesController struct {
	service *pages.PagesService
}

func NewPagesController(service *pages.PagesService) *PagesController {
	return &PagesController{service: service}
}

func (pc *PagesController) GetMetricsPage(c *gin.Context) {
	html, err := pc.service.MetricsPage()
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
