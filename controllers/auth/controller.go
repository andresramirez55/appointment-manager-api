package controller

import (
	"errors"
	"net/http"
	"os"

	"github.com/andresramirez/psych-appointments/identity"
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
		authError(c, err)
		return
	}
	setRefreshCookie(c, resp.RefreshToken, 30*24*60*60)
	resp.RefreshToken = ""
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

func (ctrl *Controller) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	resp, err := ctrl.authService.Login(c.Request.Context(), &req)
	if err != nil {
		authError(c, err)
		return
	}

	setRefreshCookie(c, resp.RefreshToken, 30*24*60*60)
	resp.RefreshToken = ""
	c.JSON(http.StatusOK, resp)
}

// Secure cross-site cookie supports separately hosted frontend/API deployments.
// Refresh and logout also require a custom header, with strict CORS in the router.
func setRefreshCookie(c *gin.Context, token string, maxAge int) {
	secure := os.Getenv("AUTH_COOKIE_SECURE") != "false"
	if secure {
		c.SetSameSite(http.SameSiteNoneMode)
	} else {
		c.SetSameSite(http.SameSiteLaxMode)
	}
	c.SetCookie("psych_refresh", token, maxAge, "/api/auth", "", secure, true)
	c.Header("Cache-Control", "no-store")
}
func (ctrl *Controller) Refresh(c *gin.Context) {
	if c.GetHeader("X-Session-Request") != "1" {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	refresh, err := c.Cookie("psych_refresh")
	if err != nil || refresh == "" {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	pair, err := ctrl.authService.Refresh(c.Request.Context(), refresh)
	if err != nil {
		if errors.Is(err, identity.ErrRejected) {
			setRefreshCookie(c, "", -1)
		}
		authError(c, err)
		return
	}
	setRefreshCookie(c, pair.RefreshToken, 30*24*60*60)
	c.JSON(http.StatusOK, gin.H{"token": pair.AccessToken})
}
func (ctrl *Controller) Logout(c *gin.Context) {
	if c.GetHeader("X-Session-Request") != "1" {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	refresh, _ := c.Cookie("psych_refresh")
	if refresh != "" {
		if err := ctrl.authService.Logout(c.Request.Context(), refresh); err != nil {
			authError(c, err)
			return
		}
	}
	setRefreshCookie(c, "", -1)
	c.Status(http.StatusNoContent)
}
func authError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, identity.ErrUnavailable):
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	case errors.Is(err, identity.ErrRejected):
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	case errors.Is(err, services.ErrAccountLinkRequired):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo completar la operación. Revisá los datos e intentá nuevamente."})
	}
}
