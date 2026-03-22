package controllers

import (
	"net/http"
	"strconv"

	"github.com/andresramirez/psych-appointments/services"
	"github.com/gin-gonic/gin"
)

type NoteController struct {
	noteService *services.NoteService
}

func NewNoteController(noteService *services.NoteService) *NoteController {
	return &NoteController{noteService: noteService}
}

func (ctrl *NoteController) Create(c *gin.Context) {
	var req services.CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	note, err := ctrl.noteService.CreateNote(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, note)
}

func (ctrl *NoteController) GetByAppointment(c *gin.Context) {
	appointmentID, err := strconv.ParseInt(c.Query("appointment_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid appointment_id"})
		return
	}

	notes, err := ctrl.noteService.GetNotesByAppointment(c.Request.Context(), appointmentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
}
