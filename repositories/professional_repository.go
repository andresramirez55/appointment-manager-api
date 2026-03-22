package repositories

import (
	"context"

	"github.com/andresramirez/psych-appointments/models"
	"gorm.io/gorm"
)

type professionalRepository struct {
	db *gorm.DB
}

func NewProfessionalRepository(db *gorm.DB) *professionalRepository {
	return &professionalRepository{db: db}
}

func (r *professionalRepository) FindByEmail(ctx context.Context, email string) (*models.Professional, error) {
	var professional models.Professional
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&professional).Error; err != nil {
		return nil, err
	}
	return &professional, nil
}
