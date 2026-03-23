package services

import (
	"context"
	"time"

	"github.com/andresramirez/psych-appointments/models"
)

type BlockRepository interface {
	Create(ctx context.Context, block *models.Block) error
	FindByProfessional(ctx context.Context, professionalID int64) ([]*models.Block, error)
	Delete(ctx context.Context, id int64) error
}

type BlockService struct {
	blockRepo BlockRepository
}

func NewBlockService(blockRepo BlockRepository) *BlockService {
	return &BlockService{blockRepo: blockRepo}
}

type CreateBlockRequest struct {
	ProfessionalID int64     `json:"professional_id"`
	StartsAt       time.Time `json:"starts_at"`
	EndsAt         time.Time `json:"ends_at"`
	Reason         string    `json:"reason"`
}

func (s *BlockService) CreateBlock(ctx context.Context, req *CreateBlockRequest) (*models.Block, error) {
	block := &models.Block{
		ProfessionalID: req.ProfessionalID,
		StartsAt:       req.StartsAt,
		EndsAt:         req.EndsAt,
		Reason:         req.Reason,
	}
	if err := s.blockRepo.Create(ctx, block); err != nil {
		return nil, err
	}
	return block, nil
}

func (s *BlockService) GetBlocks(ctx context.Context, professionalID int64) ([]*models.Block, error) {
	return s.blockRepo.FindByProfessional(ctx, professionalID)
}

func (s *BlockService) DeleteBlock(ctx context.Context, id int64) error {
	return s.blockRepo.Delete(ctx, id)
}
