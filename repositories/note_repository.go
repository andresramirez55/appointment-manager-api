package repositories

import (
	"context"

	"github.com/andresramirez/psych-appointments/models"
	"gorm.io/gorm"
)

type noteRepository struct {
	db *gorm.DB
}

func NewNoteRepository(db *gorm.DB) *noteRepository {
	return &noteRepository{db: db}
}

func (r *noteRepository) Create(ctx context.Context, note *models.SessionNote) error {
	return r.db.WithContext(ctx).Create(note).Error
}

func (r *noteRepository) FindByAppointmentID(ctx context.Context, appointmentID int64) ([]*models.SessionNote, error) {
	var notes []*models.SessionNote
	if err := r.db.WithContext(ctx).
		Where("appointment_id = ?", appointmentID).
		Order("created_at DESC").
		Find(&notes).Error; err != nil {
		return nil, err
	}
	return notes, nil
}
