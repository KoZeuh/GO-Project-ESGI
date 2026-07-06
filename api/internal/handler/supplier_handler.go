package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

// SupplierHandler expose les routes CRUD sous /api/v1/suppliers.
type SupplierHandler struct {
	suppliers *service.SupplierService
}

func NewSupplierHandler(suppliers *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{suppliers: suppliers}
}

func (h *SupplierHandler) List(c *gin.Context) {
	suppliers, err := h.suppliers.List()
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, suppliers)
}

func (h *SupplierHandler) Get(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	supplier, err := h.suppliers.Get(id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, supplier)
}

func (h *SupplierHandler) Create(c *gin.Context) {
	var req dto.SupplierRequest
	if !bindJSON(c, &req) {
		return
	}
	supplier, err := h.suppliers.Create(req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, supplier)
}

func (h *SupplierHandler) Update(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var req dto.SupplierRequest
	if !bindJSON(c, &req) {
		return
	}
	supplier, err := h.suppliers.Update(id, req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, supplier)
}

func (h *SupplierHandler) Delete(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if err := h.suppliers.Delete(id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
