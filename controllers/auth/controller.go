package controller

import (
	"net/http"

	"github.com/andresramirez/psych-appointments/services"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	authService *services.AuthService
}

func New(authService *services.AuthService) *Controller {
	return &Controller{authService: authService}
}

func (ctrl *Controller) Register(c *gin.Context) {
	var req services.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if req.Email == "" || req.Password == "" || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email, password and name are required"})
		return
	}

	resp, err := ctrl.authService.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (ctrl *Controller) GetProfile(c *gin.Context) {
	professionalID := c.MustGet("professional_id").(int64)
	professional, err := ctrl.authService.GetProfile(c.Request.Context(), professionalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Profile not found"})
		return
	}
	c.JSON(http.StatusOK, professional)
}

func (ctrl *Controller) UpdateProfile(c *gin.Context) {
	professionalID := c.MustGet("professional_id").(int64)
	var req services.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	professional, err := ctrl.authService.UpdateProfile(c.Request.Context(), professionalID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, professional)
}

func (ctrl *Controller) UpdatePassword(c *gin.Context) {
	professionalID := c.MustGet("professional_id").(int64)
	var req services.UpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	if err := ctrl.authService.UpdatePassword(c.Request.Context(), professionalID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Contraseña actualizada"})
}

func (ctrl *Controller) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	resp, err := ctrl.authService.Login(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
