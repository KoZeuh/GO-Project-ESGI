package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/middleware"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

// Handlers regroupe tous les contrôleurs nécessaires au montage des routes.
type Handlers struct {
	Auth      *AuthHandler
	Products  *ProductHandler
	Suppliers *SupplierHandler
	Movements *MovementHandler
	Alerts    *AlertHandler
	Export    *ExportHandler
}

// NewRouter construit le moteur Gin et déclare toutes les routes de l'API.
func NewRouter(h Handlers, authService *service.AuthService) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.POST("/register", h.Auth.Register)
	auth.POST("/login", h.Auth.Login)

	protected := v1.Group("")
	protected.Use(middleware.Auth(authService))

	products := protected.Group("/products")
	products.GET("", h.Products.List)
	products.GET("/:id", h.Products.Get)
	products.POST("", h.Products.Create)
	products.PUT("/:id", h.Products.Update)
	products.DELETE("/:id", h.Products.Delete)

	suppliers := protected.Group("/suppliers")
	suppliers.GET("", h.Suppliers.List)
	suppliers.GET("/:id", h.Suppliers.Get)
	suppliers.POST("", h.Suppliers.Create)
	suppliers.PUT("/:id", h.Suppliers.Update)
	suppliers.DELETE("/:id", h.Suppliers.Delete)

	movements := protected.Group("/movements")
	movements.GET("", h.Movements.List)
	movements.POST("", h.Movements.Create)

	alerts := protected.Group("/alerts")
	alerts.GET("", h.Alerts.List)
	alerts.PUT("/settings", h.Alerts.UpdateSettings)

	protected.GET("/export", h.Export.Export)

	return r
}
