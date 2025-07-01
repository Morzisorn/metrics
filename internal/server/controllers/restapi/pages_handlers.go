package restapi

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/morzisorn/metrics/internal/server/services/pages"
)

type PagesController struct {
	service *pages.PagesService
}

// NewPagesController receives Pages service and creates Pages controllers, returns pointer.
func NewPagesController(service *pages.PagesService) *PagesController {
	return &PagesController{service: service}
}

// GetMetricsPage return html page with all metrics
func (pc *PagesController) GetMetricsPage(c *gin.Context) {
	html, err := pc.service.MetricsPage()
	if err != nil {
		c.Status(http.StatusInternalServerError)
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}
