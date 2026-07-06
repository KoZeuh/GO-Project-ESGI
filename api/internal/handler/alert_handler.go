package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

// AlertHandler expose GET /api/v1/alerts et PUT /api/v1/alerts/settings.
type AlertHandler struct {
	alerts *service.AlertService
}

func NewAlertHandler(alerts *service.AlertService) *AlertHandler {
	return &AlertHandler{alerts: alerts}
}

// List gère GET /api/v1/alerts : seuil par défaut + produits en stock faible.
func (h *AlertHandler) List(c *gin.Context) {
	defaultThreshold, lowStock, err := h.alerts.GetLowStockProducts()
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.AlertsResponse{
		DefaultThreshold: defaultThreshold,
		LowStockProducts: lowStock,
	})
}

// UpdateSettings gère PUT /api/v1/alerts/settings.
func (h *AlertHandler) UpdateSettings(c *gin.Context) {
	var req dto.AlertSettingsRequest
	if !bindJSON(c, &req) {
		return
	}

	settings, err := h.alerts.UpdateSettings(req.DefaultThreshold)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, settings)
}
