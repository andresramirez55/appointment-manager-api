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

func (s *PatientService) GetAllPatients(ctx context.Context, professionalID int64, consultorioID *int64) ([]*models.Patient, error) {
	return s.patientRepo.FindAll(ctx, professionalID, consultorioID)
}

type CreatePatientRequest struct {
	Name          string `json:"name"`
	Phone         string `json:"phone"`
	Email         string `json:"email"`
	Notes         string `json:"notes"`
	ConsultorioID *int64 `json:"consultorio_id"`
}

func (s *PatientService) UpdatePatient(ctx context.Context, id int64, req *CreatePatientRequest) (*models.Patient, error) {
	patient, err := s.patientRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	patient.Name = req.Name
	patient.Phone = req.Phone
	patient.Email = req.Email
	patient.Notes = req.Notes
	if req.ConsultorioID != nil {
		patient.ConsultorioID = req.ConsultorioID
	}
	if err := s.patientRepo.Update(ctx, patient); err != nil {
		return nil, err
	}
	return patient, nil
}

func (s *PatientService) CreatePatient(ctx context.Context, professionalID int64, req *CreatePatientRequest) (*models.Patient, error) {
	patient := &models.Patient{
		ProfessionalID: professionalID,
		ConsultorioID:  req.ConsultorioID,
		Name:           req.Name,
		Phone:          req.Phone,
		Email:          req.Email,
		Notes:          req.Notes,
	}
	if err := s.patientRepo.Create(ctx, patient); err != nil {
		return nil, err
	}
	return patient, nil
}
