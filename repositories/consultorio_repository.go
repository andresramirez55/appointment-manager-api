package repositories

import (
	"context"

	"github.com/andresramirez/psych-appointments/models"
	"gorm.io/gorm"
)

type consultorioRepository struct {
	db *gorm.DB
}

func NewConsultorioRepository(db *gorm.DB) *consultorioRepository {
	return &consultorioRepository{db: db}
}

func (r *consultorioRepository) Create(ctx context.Context, c *models.Consultorio) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *consultorioRepository) FindAll(ctx context.Context, professionalID int64) ([]*models.Consultorio, error) {
	var consultorios []*models.Consultorio
	if err := r.db.WithContext(ctx).Where("professional_id = ?", professionalID).Order("id ASC").Find(&consultorios).Error; err != nil {
		return nil, err
	}
	return consultorios, nil
}

func (r *consultorioRepository) FindByID(ctx context.Context, id int64) (*models.Consultorio, error) {
	var c models.Consultorio
	if err := r.db.WithContext(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *consultorioRepository) Update(ctx context.Context, c *models.Consultorio) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *consultorioRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Consultorio{}, id).Error
}
