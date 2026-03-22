package controllers

import (
	"net/http"
	"strconv"

	"github.com/andresramirez/psych-appointments/services"
	"github.com/gin-gonic/gin"
)

type AvailabilityController struct {
	availabilityService *services.AvailabilityService
}

func NewAvailabilityController(availabilityService *services.AvailabilityService) *AvailabilityController {
	return &AvailabilityController{availabilityService: availabilityService}
}

func (ctrl *AvailabilityController) CreateSlot(c *gin.Context) {
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

func (ctrl *AvailabilityController) GetSlots(c *gin.Context) {
	professionalID := c.GetInt64("professional_id")

	slots, err := ctrl.availabilityService.GetSlots(c.Request.Context(), professionalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, slots)
}

func (ctrl *AvailabilityController) DeleteSlot(c *gin.Context) {
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

func (ctrl *AvailabilityController) CreateOverride(c *gin.Context) {
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
