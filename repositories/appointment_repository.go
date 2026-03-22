package repositories

import (
	"context"
	"time"

	"github.com/andresramirez/psych-appointments/models"
	"gorm.io/gorm"
)

type appointmentRepository struct {
	db *gorm.DB
}

func NewAppointmentRepository(db *gorm.DB) *appointmentRepository {
	return &appointmentRepository{db: db}
}

func (r *appointmentRepository) Create(ctx context.Context, appointment *models.Appointment) error {
	return r.db.WithContext(ctx).Create(appointment).Error
}

func (r *appointmentRepository) FindByID(ctx context.Context, id int64) (*models.Appointment, error) {
	var appointment models.Appointment
	if err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Professional").
		First(&appointment, id).Error; err != nil {
		return nil, err
	}
	return &appointment, nil
}

func (r *appointmentRepository) FindAll(ctx context.Context) ([]*models.Appointment, error) {
	var appointments []*models.Appointment
	if err := r.db.WithContext(ctx).
		Preload("Patient").
		Preload("Professional").
		Order("starts_at DESC").
		Find(&appointments).Error; err != nil {
		return nil, err
	}
	return appointments, nil
}

func (r *appointmentRepository) Update(ctx context.Context, appointment *models.Appointment) error {
	return r.db.WithContext(ctx).Save(appointment).Error
}

func (r *appointmentRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Appointment{}, id).Error
}

func (r *appointmentRepository) FindByPatient(ctx context.Context, patientID int64) ([]*models.Appointment, error) {
	var appointments []*models.Appointment
	if err := r.db.WithContext(ctx).
		Preload("Patient").
		Where("patient_id = ?", patientID).
		Order("starts_at DESC").
		Find(&appointments).Error; err != nil {
		return nil, err
	}
	return appointments, nil
}

func (r *appointmentRepository) FindPendingReminders(ctx context.Context, from, to time.Time) ([]*models.Appointment, error) {
	var appointments []*models.Appointment
	if err := r.db.WithContext(ctx).
		Preload("Patient").
		Where("starts_at BETWEEN ? AND ?", from, to).
		Where("reminder_sent_at IS NULL").
		Where("status = ?", "scheduled").
		Find(&appointments).Error; err != nil {
		return nil, err
	}
	return appointments, nil
}
