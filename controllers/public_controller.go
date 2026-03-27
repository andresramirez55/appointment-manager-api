package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/andresramirez/psych-appointments/services"
	"github.com/gin-gonic/gin"
)

type PublicController struct {
	availabilityService *services.AvailabilityService
	appointmentService  *services.AppointmentService
	authService         *services.AuthService
}

func NewPublicController(
	availabilityService *services.AvailabilityService,
	appointmentService *services.AppointmentService,
	authService *services.AuthService,
) *PublicController {
	return &PublicController{
		availabilityService: availabilityService,
		appointmentService:  appointmentService,
		authService:         authService,
	}
}

func (ctrl *PublicController) GetProfessional(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	professional, err := ctrl.authService.GetProfile(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Professional not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":        professional.ID,
		"name":      professional.Name,
		"specialty": professional.Specialty,
	})
}

func (ctrl *PublicController) GetAvailableSlots(c *gin.Context) {
	professionalID, err := strconv.ParseInt(c.Query("professional_id"), 10, 64)
	if err != nil || professionalID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "professional_id required"})
		return
	}

	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date parameter required (YYYY-MM-DD)"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format"})
		return
	}

	slots, err := ctrl.availabilityService.GetAvailableSlots(c.Request.Context(), professionalID, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slots)
}

func (ctrl *PublicController) GetAppointmentByToken(c *gin.Context) {
	token := c.Param("token")
	appointment, err := ctrl.appointmentService.GetByCancelToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Turno no encontrado"})
		return
	}
	patientName := ""
	if appointment.Patient != nil {
		patientName = appointment.Patient.Name
	}
	professionalName := ""
	if appointment.Professional != nil {
		professionalName = appointment.Professional.Name
	}
	c.JSON(http.StatusOK, gin.H{
		"patient_name":      patientName,
		"professional_name": professionalName,
		"starts_at":         appointment.StartsAt,
		"duration_minutes":  appointment.DurationMinutes,
		"status":            appointment.Status,
	})
}

func (ctrl *PublicController) CancelByToken(c *gin.Context) {
	token := c.Param("token")
	if err := ctrl.appointmentService.CancelByToken(c.Request.Context(), token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Turno cancelado"})
}

func (ctrl *PublicController) CreateAppointment(c *gin.Context) {
	var req services.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.ProfessionalID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "professional_id required"})
		return
	}

	appointment, err := ctrl.appointmentService.CreateAppointment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, appointment)
}
