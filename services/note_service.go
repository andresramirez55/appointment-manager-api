package services

import (
	"context"
	"fmt"

	"github.com/andresramirez/psych-appointments/models"
)

// NoteRepository define métodos para acceso a notas
type NoteRepository interface {
	Create(ctx context.Context, note *models.SessionNote) error
	FindByAppointmentID(ctx context.Context, appointmentID int64) ([]*models.SessionNote, error)
}

// NoteService maneja lógica de notas de sesión
type NoteService struct {
	noteRepo NoteRepository
}

func NewNoteService(noteRepo NoteRepository) *NoteService {
	return &NoteService{
		noteRepo: noteRepo,
	}
}

type CreateNoteRequest struct {
	AppointmentID int64  `json:"appointment_id"`
	Content       string `json:"content"`
}

func (s *NoteService) CreateNote(ctx context.Context, req *CreateNoteRequest) (*models.SessionNote, error) {
	note := &models.SessionNote{
		AppointmentID: req.AppointmentID,
		Content:       req.Content,
	}

	if err := s.noteRepo.Create(ctx, note); err != nil {
		return nil, fmt.Errorf("failed to create note: %w", err)
	}

	return note, nil
}

func (s *NoteService) GetNotesByAppointment(ctx context.Context, appointmentID int64) ([]*models.SessionNote, error) {
	return s.noteRepo.FindByAppointmentID(ctx, appointmentID)
}
