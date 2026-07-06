package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/KoZeuh/GO-Project-ESGI/api/internal/dto"
	"github.com/KoZeuh/GO-Project-ESGI/api/internal/service"
)

// AuthHandler expose les routes POST /auth/register et POST /auth/login.
type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// Register gère POST /api/v1/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !bindJSON(c, &req) {
		return
	}

	user, err := h.auth.Register(req.Username, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, dto.NewUserResponse(user))
}

// Login gère POST /api/v1/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, &req) {
		return
	}

	token, expiresAt, user, err := h.auth.Login(req.Username, req.Password)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      dto.NewUserResponse(user),
	})
}
