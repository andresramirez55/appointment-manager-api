package controllers

import (
	"net/http"
	"time"

	"github.com/andresramirez/psych-appointments/services"
	"github.com/gin-gonic/gin"
)

type PublicController struct {
	availabilityService *services.AvailabilityService
	appointmentService  *services.AppointmentService
}

func NewPublicController(
	availabilityService *services.AvailabilityService,
	appointmentService *services.AppointmentService,
) *PublicController {
	return &PublicController{
		availabilityService: availabilityService,
		appointmentService:  appointmentService,
	}
}

func (ctrl *PublicController) GetAvailableSlots(c *gin.Context) {
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

	// Por ahora, asumimos professional_id = 1 (único profesional)
	slots, err := ctrl.availabilityService.GetAvailableSlots(c.Request.Context(), 1, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slots)
}

func (ctrl *PublicController) CreateAppointment(c *gin.Context) {
	var req services.CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Por ahora, asumimos professional_id = 1 (único profesional)
	req.ProfessionalID = 1

	appointment, err := ctrl.appointmentService.CreateAppointment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, appointment)
}
