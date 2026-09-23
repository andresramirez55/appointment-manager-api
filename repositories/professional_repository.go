package repositories

import (
	"context"
	"errors"
	"github.com/andresramirez/psych-appointments/domain"

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
	if err := r.db.WithContext(ctx).Where("LOWER(email) = LOWER(?)", email).First(&professional).Error; err != nil {
		return nil, professionalError(err)
	}
	return &professional, nil
}

func (r *professionalRepository) Create(ctx context.Context, professional *models.Professional) error {
	return r.db.WithContext(ctx).Create(professional).Error
}

func (r *professionalRepository) FindByID(ctx context.Context, id int64) (*models.Professional, error) {
	var professional models.Professional
	if err := r.db.WithContext(ctx).First(&professional, id).Error; err != nil {
		return nil, professionalError(err)
	}
	return &professional, nil
}

func (r *professionalRepository) Update(ctx context.Context, professional *models.Professional) error {
	return r.db.WithContext(ctx).Save(professional).Error
}

func professionalError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrProfessionalNotFound
	}
	return err
}
func (r *professionalRepository) FindByAuthUserID(ctx context.Context, id string) (*models.Professional, error) {
	var p models.Professional
	if err := r.db.WithContext(ctx).Where("auth_user_id = ?", id).First(&p).Error; err != nil {
		return nil, professionalError(err)
	}
	return &p, nil
}
