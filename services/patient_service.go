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
