package repositories

import (
	"context"

	"github.com/andresramirez/psych-appointments/models"
	"gorm.io/gorm"
)

type blockRepository struct {
	db *gorm.DB
}

func NewBlockRepository(db *gorm.DB) *blockRepository {
	return &blockRepository{db: db}
}

func (r *blockRepository) Create(ctx context.Context, block *models.Block) error {
	return r.db.WithContext(ctx).Create(block).Error
}

func (r *blockRepository) FindByProfessional(ctx context.Context, professionalID int64) ([]*models.Block, error) {
	var blocks []*models.Block
	if err := r.db.WithContext(ctx).
		Where("professional_id = ?", professionalID).
		Order("starts_at ASC").
		Find(&blocks).Error; err != nil {
		return nil, err
	}
	return blocks, nil
}

func (r *blockRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&models.Block{}, id).Error
}
