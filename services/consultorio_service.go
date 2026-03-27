package services

import (
	"context"

	"github.com/andresramirez/psych-appointments/models"
)

type ConsultorioRepository interface {
	Create(ctx context.Context, c *models.Consultorio) error
	FindAll(ctx context.Context, professionalID int64) ([]*models.Consultorio, error)
	FindByID(ctx context.Context, id int64) (*models.Consultorio, error)
	Update(ctx context.Context, c *models.Consultorio) error
	Delete(ctx context.Context, id int64) error
}

type ConsultorioService struct {
	repo ConsultorioRepository
}

func NewConsultorioService(repo ConsultorioRepository) *ConsultorioService {
	return &ConsultorioService{repo: repo}
}

type CreateConsultorioRequest struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

func (s *ConsultorioService) GetAll(ctx context.Context, professionalID int64) ([]*models.Consultorio, error) {
	return s.repo.FindAll(ctx, professionalID)
}

func (s *ConsultorioService) Create(ctx context.Context, professionalID int64, req *CreateConsultorioRequest) (*models.Consultorio, error) {
	c := &models.Consultorio{
		ProfessionalID: professionalID,
		Name:           req.Name,
		Address:        req.Address,
	}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ConsultorioService) Update(ctx context.Context, id int64, req *CreateConsultorioRequest) (*models.Consultorio, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	c.Name = req.Name
	c.Address = req.Address
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, err
	}
	return c, nil
}

func (s *ConsultorioService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
