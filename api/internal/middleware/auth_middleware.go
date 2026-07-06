// Package middleware contient les middlewares HTTP partagés par les routes de l'API.

package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

// Clés utilisées pour stocker les informations de l'utilisateur authentifié dans le gin.Context, renseignées par Auth() et lues par les handlers.
const (
	ContextUserID   = "user_id"
	ContextUsername = "username"
	ContextRole     = "role"
)

// Auth vérifie la présence et la validité d'un JWT dans l'en-tête "Authorization: Bearer <token>" et rejette la requête avec 401 sinon.
func Auth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" || parts[1] == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "en-tête Authorization manquant ou mal formé"})
			return
		}

		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "token invalide ou expiré"})
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextUsername, claims.Username)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}
