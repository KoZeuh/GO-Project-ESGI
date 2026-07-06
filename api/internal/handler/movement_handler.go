package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/middleware"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/models"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

// MovementHandler expose GET/POST /api/v1/movements.
type MovementHandler struct {
	movements *service.MovementService
}

func NewMovementHandler(movements *service.MovementService) *MovementHandler {
	return &MovementHandler{movements: movements}
}

// List gère GET /api/v1/movements?product_id=&type=&from=&to=
func (h *MovementHandler) List(c *gin.Context) {
	filter := repository.MovementFilter{
		Type: models.MovementType(c.Query("type")),
		From: c.Query("from"),
		To:   c.Query("to"),
	}
	if raw := c.Query("product_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "product_id invalide"})
			return
		}
		filter.ProductID = id
	}

	movements, err := h.movements.List(filter)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, movements)
}

// Create gère POST /api/v1/movements. L'utilisateur est déduit du JWT (middleware.Auth).
func (h *MovementHandler) Create(c *gin.Context) {
	var req dto.MovementRequest
	if !bindJSON(c, &req) {
		return
	}

	userID := c.GetInt64(middleware.ContextUserID)
	movement, err := h.movements.Create(req, userID)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, movement)
}
