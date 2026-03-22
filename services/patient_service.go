package services

import (
	"context"

	"github.com/andresramirez/psych-appointments/models"
)

// PatientService maneja lógica de pacientes
type PatientService struct {
	patientRepo PatientRepository
}

func NewPatientService(patientRepo PatientRepository) *PatientService {
	return &PatientService{
		patientRepo: patientRepo,
	}
}

func (s *PatientService) GetPatient(ctx context.Context, id int64) (*models.Patient, error) {
	return s.patientRepo.FindByID(ctx, id)
}

func (s *PatientService) GetAllPatients(ctx context.Context) ([]*models.Patient, error) {
	return s.patientRepo.FindAll(ctx)
}

type CreatePatientRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Notes string `json:"notes"`
}

func (s *PatientService) CreatePatient(ctx context.Context, req *CreatePatientRequest) (*models.Patient, error) {
	patient := &models.Patient{
		Name:  req.Name,
		Phone: req.Phone,
		Email: req.Email,
		Notes: req.Notes,
	}
	if err := s.patientRepo.Create(ctx, patient); err != nil {
		return nil, err
	}
	return patient, nil
}
