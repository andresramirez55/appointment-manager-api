package repositories

import (
	"context"
	"time"

	"github.com/andresramirez/psych-appointments/models"
	"gorm.io/gorm"
)

type availabilityRepository struct {
	db *gorm.DB
}

func NewAvailabilityRepository(db *gorm.DB) *availabilityRepository {
	return &availabilityRepository{db: db}
}

// Slots regulares
func (r *availabilityRepository) CreateSlot(ctx context.Context, slot *models.AvailabilitySlot) error {
	return r.db.WithContext(ctx).Create(slot).Error
}

func (r *availabilityRepository) FindSlotsByProfessional(ctx context.Context, professionalID int64) ([]*models.AvailabilitySlot, error) {
	var slots []*models.AvailabilitySlot
	if err := r.db.WithContext(ctx).
		Where("professional_id = ?", professionalID).
		Order("day_of_week ASC, start_time ASC").
		Find(&slots).Error; err != nil {
		return nil, err
	}
	return slots, nil
}

func (r *availabilityRepository) UpdateSlot(ctx context.Context, slot *models.AvailabilitySlot) error {
	return r.db.WithContext(ctx).Save(slot).Error
}

func (r *availabilityRepository) DeleteSlot(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.AvailabilitySlot{}, id).Error
}

// Overrides (excepciones)
func (r *availabilityRepository) CreateOverride(ctx context.Context, override *models.AvailabilityOverride) error {
	return r.db.WithContext(ctx).Create(override).Error
}

func (r *availabilityRepository) FindOverridesByProfessional(ctx context.Context, professionalID int64) ([]*models.AvailabilityOverride, error) {
	var overrides []*models.AvailabilityOverride
	if err := r.db.WithContext(ctx).
		Where("professional_id = ?", professionalID).
		Order("date ASC").
		Find(&overrides).Error; err != nil {
		return nil, err
	}
	return overrides, nil
}

func (r *availabilityRepository) FindOverrideByDate(ctx context.Context, professionalID int64, date time.Time) (*models.AvailabilityOverride, error) {
	var override models.AvailabilityOverride
	// Buscar override para la fecha específica (ignorando horas)
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	if err := r.db.WithContext(ctx).
		Where("professional_id = ?", professionalID).
		Where("date >= ? AND date < ?", startOfDay, endOfDay).
		First(&override).Error; err != nil {
		return nil, err
	}
	return &override, nil
}
