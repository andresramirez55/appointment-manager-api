package controller

import (
	"net/http"
	"strconv"

	"github.com/andresramirez/psych-appointments/services"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	blockService *services.BlockService
}

func New(blockService *services.BlockService) *Controller {
	return &Controller{blockService: blockService}
}

func (ctrl *Controller) Create(c *gin.Context) {
	var req services.CreateBlockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	req.ProfessionalID = c.MustGet("professional_id").(int64)

	block, err := ctrl.blockService.CreateBlock(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, block)
}

func (ctrl *Controller) GetAll(c *gin.Context) {
	professionalID := c.MustGet("professional_id").(int64)
	blocks, err := ctrl.blockService.GetBlocks(c.Request.Context(), professionalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, blocks)
}

func (ctrl *Controller) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	if err := ctrl.blockService.DeleteBlock(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Block deleted"})
}
