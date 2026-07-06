package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

// ProductHandler expose les routes CRUD sous /api/v1/products.
type ProductHandler struct {
	products *service.ProductService
}

func NewProductHandler(products *service.ProductService) *ProductHandler {
	return &ProductHandler{products: products}
}

// List gère GET /api/v1/products?search=...
func (h *ProductHandler) List(c *gin.Context) {
	products, err := h.products.List(c.Query("search"))
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, products)
}

// Get gère GET /api/v1/products/:id
func (h *ProductHandler) Get(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	product, err := h.products.Get(id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, product)
}

// Create gère POST /api/v1/products
func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.ProductRequest
	if !bindJSON(c, &req) {
		return
	}
	product, err := h.products.Create(req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, product)
}

// Update gère PUT /api/v1/products/:id
func (h *ProductHandler) Update(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	var req dto.ProductRequest
	if !bindJSON(c, &req) {
		return
	}
	product, err := h.products.Update(id, req)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, product)
}

// Delete gère DELETE /api/v1/products/:id
func (h *ProductHandler) Delete(c *gin.Context) {
	id, ok := paramID(c, "id")
	if !ok {
		return
	}
	if err := h.products.Delete(id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
