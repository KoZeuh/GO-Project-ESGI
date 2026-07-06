package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

// ExportHandler expose GET /api/v1/export : un instantané complet du stock, consommé par le TUI pour son import et son cache local.
type ExportHandler struct {
	products  *service.ProductService
	suppliers *service.SupplierService
	movements *service.MovementService
}

func NewExportHandler(products *service.ProductService, suppliers *service.SupplierService, movements *service.MovementService) *ExportHandler {
	return &ExportHandler{products: products, suppliers: suppliers, movements: movements}
}

func (h *ExportHandler) Export(c *gin.Context) {
	products, err := h.products.List("")
	if err != nil {
		respondError(c, err)
		return
	}
	suppliers, err := h.suppliers.List()
	if err != nil {
		respondError(c, err)
		return
	}
	movements, err := h.movements.List(repository.MovementFilter{})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.ExportResponse{
		ExportedAt: time.Now(),
		Products:   products,
		Suppliers:  suppliers,
		Movements:  movements,
	})
}
