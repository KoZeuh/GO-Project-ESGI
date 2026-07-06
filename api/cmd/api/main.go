// Point d'entrée du serveur API : charge la configuration, connecte la base de données, assemble les couches (repository -> service -> handler) et démarre le serveur HTTP.

package main

import (
	"log"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/config"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/database"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/handler"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("connexion base de données : %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	supplierRepo := repository.NewSupplierRepository(db)
	productRepo := repository.NewProductRepository(db)
	movementRepo := repository.NewMovementRepository(db)
	alertRepo := repository.NewAlertRepository(db)

	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiration)
	alertService := service.NewAlertService(alertRepo, productRepo)
	productService := service.NewProductService(productRepo, supplierRepo, alertService)
	supplierService := service.NewSupplierService(supplierRepo)
	movementService := service.NewMovementService(db, productRepo, movementRepo)

	handlers := handler.Handlers{
		Auth:      handler.NewAuthHandler(authService),
		Products:  handler.NewProductHandler(productService),
		Suppliers: handler.NewSupplierHandler(supplierService),
		Movements: handler.NewMovementHandler(movementService),
		Alerts:    handler.NewAlertHandler(alertService),
		Export:    handler.NewExportHandler(productService, supplierService, movementService),
	}

	router := handler.NewRouter(handlers, authService)

	log.Printf("API démarrée sur http://localhost:%s (env=%s, db=%s)", cfg.Port, cfg.Env, cfg.DBPath)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("arrêt du serveur : %v", err)
	}
}
