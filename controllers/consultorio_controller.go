package controllers

import (
	"net/http"
	"strconv"

	"github.com/andresramirez/psych-appointments/services"
	"github.com/gin-gonic/gin"
)

type ConsultorioController struct {
	service *services.ConsultorioService
}

func NewConsultorioController(service *services.ConsultorioService) *ConsultorioController {
	return &ConsultorioController{service: service}
}

func (ctrl *ConsultorioController) GetAll(c *gin.Context) {
	professionalID := c.MustGet("professional_id").(int64)
	consultorios, err := ctrl.service.GetAll(c.Request.Context(), professionalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, consultorios)
}

func (ctrl *ConsultorioController) Create(c *gin.Context) {
	var req services.CreateConsultorioRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	professionalID := c.MustGet("professional_id").(int64)
	consultorio, err := ctrl.service.Create(c.Request.Context(), professionalID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, consultorio)
}

func (ctrl *ConsultorioController) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	var req services.CreateConsultorioRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	consultorio, err := ctrl.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, consultorio)
}

func (ctrl *ConsultorioController) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := ctrl.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Consultorio deleted"})
}
