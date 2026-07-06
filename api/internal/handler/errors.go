// Package handler contient les contrôleurs HTTP (couche la plus externe) : ils décodent la requête, appellent la couche service, et sérialisent la réponse.
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/repository"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

// respondError traduit une erreur de la couche service/repository en réponse HTTP adaptée, afin que chaque handler n'ait pas à connaître ce mapping.
func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, repository.ErrDuplicate):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "une ressource avec ces informations existe déjà"})
	case errors.Is(err, repository.ErrInsufficient):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "quantité en stock insuffisante pour cette sortie"})
	case errors.Is(err, service.ErrSupplierNotFound):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "erreur interne du serveur"})
	}
}

// bindJSON décode et valide le corps JSON de la requête ; en cas d'erreur, elle répond directement en 400 et retourne false pour que l'appelant s'arrête.
func bindJSON(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return false
	}
	return true
}

// paramID extrait et valide un identifiant numérique depuis l'URL (ex: /products/:id).
func paramID(c *gin.Context, name string) (int64, bool) {
	id, err := parseID(c.Param(name))
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "identifiant invalide"})
		return 0, false
	}
	return id, true
}
