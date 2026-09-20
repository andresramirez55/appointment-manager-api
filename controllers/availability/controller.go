package controller

import (
	"net/http"
	"strconv"

	"github.com/andresramirez/psych-appointments/services"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	availabilityService *services.AvailabilityService
}

func New(availabilityService *services.AvailabilityService) *Controller {
	return &Controller{availabilityService: availabilityService}
}

func (ctrl *Controller) CreateSlot(c *gin.Context) {
	professionalID := c.GetInt64("professional_id")

	var req services.CreateSlotRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	req.ProfessionalID = professionalID

	slot, err := ctrl.availabilityService.CreateSlot(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, slot)
}

func (ctrl *Controller) GetSlots(c *gin.Context) {
	professionalID := c.GetInt64("professional_id")

	slots, err := ctrl.availabilityService.GetSlots(c.Request.Context(), professionalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slots)
}

func (ctrl *Controller) DeleteSlot(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	if err := ctrl.availabilityService.DeleteSlot(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Slot deleted"})
}

func (ctrl *Controller) CreateOverride(c *gin.Context) {
	professionalID := c.GetInt64("professional_id")

	var req services.CreateOverrideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	req.ProfessionalID = professionalID

	override, err := ctrl.availabilityService.CreateOverride(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, override)
}
