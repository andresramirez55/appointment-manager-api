package services

import (
	"context"
	"fmt"
	"time"

	"github.com/andresramirez/psych-appointments/models"
)

// AppointmentRepository define métodos para acceso a turnos
type AppointmentRepository interface {
	Create(ctx context.Context, appointment *models.Appointment) error
	FindByID(ctx context.Context, id int64) (*models.Appointment, error)
	FindAll(ctx context.Context) ([]*models.Appointment, error)
	Update(ctx context.Context, appointment *models.Appointment) error
	Delete(ctx context.Context, id int64) error
	FindPendingReminders(ctx context.Context, from, to time.Time) ([]*models.Appointment, error)
}

// PatientRepository define métodos para acceso a pacientes
type PatientRepository interface {
	Create(ctx context.Context, patient *models.Patient) error
	FindByPhone(ctx context.Context, phone string) (*models.Patient, error)
	FindByID(ctx context.Context, id int64) (*models.Patient, error)
	FindAll(ctx context.Context) ([]*models.Patient, error)
}

// AppointmentService maneja lógica de turnos
type AppointmentService struct {
	appointmentRepo AppointmentRepository
	patientRepo     PatientRepository
	whatsappSender  WhatsAppSender
	emailService    *EmailService
}

func NewAppointmentService(
	appointmentRepo AppointmentRepository,
	patientRepo PatientRepository,
	whatsappSender WhatsAppSender,
	emailService *EmailService,
) *AppointmentService {
	return &AppointmentService{
		appointmentRepo: appointmentRepo,
		patientRepo:     patientRepo,
		whatsappSender:  whatsappSender,
		emailService:    emailService,
	}
}

type CreateAppointmentRequest struct {
	PatientName     string    `json:"patient_name"`
	PatientPhone    string    `json:"patient_phone"`
	ProfessionalID  int64     `json:"professional_id"`
	StartsAt        time.Time `json:"starts_at"`
	DurationMinutes int       `json:"duration_minutes"`
}

func (s *AppointmentService) CreateAppointment(ctx context.Context, req *CreateAppointmentRequest) (*models.Appointment, error) {
	// Buscar o crear paciente
	patient, err := s.patientRepo.FindByPhone(ctx, req.PatientPhone)
	if err != nil {
		// Paciente no existe, crear uno nuevo
		patient = &models.Patient{
			Name:  req.PatientName,
			Phone: req.PatientPhone,
		}
		if err := s.patientRepo.Create(ctx, patient); err != nil {
			return nil, fmt.Errorf("failed to create patient: %w", err)
		}
	}

	// Crear turno
	appointment := &models.Appointment{
		PatientID:       patient.ID,
		ProfessionalID:  req.ProfessionalID,
		StartsAt:        req.StartsAt,
		DurationMinutes: req.DurationMinutes,
		Status:          "scheduled",
	}

	if err := s.appointmentRepo.Create(ctx, appointment); err != nil {
		return nil, fmt.Errorf("failed to create appointment: %w", err)
	}

	// Cargar relaciones
	appointment.Patient = patient

	// Enviar confirmación por WhatsApp
	message := fmt.Sprintf(
		"✅ Turno confirmado\n\nFecha: %s\nDuración: %d minutos\n\nGracias por reservar.",
		appointment.StartsAt.Format("02/01/2006 15:04"),
		appointment.DurationMinutes,
	)
	if err := s.whatsappSender.SendMessage(ctx, patient.Phone, message); err != nil {
		// Log error pero no fallar la creación del turno
		fmt.Printf("Warning: failed to send WhatsApp confirmation: %v\n", err)
	}

	return appointment, nil
}

type CreateAppointmentByPatientRequest struct {
	PatientID       int64     `json:"patient_id"`
	ProfessionalID  int64     `json:"professional_id"`
	StartsAt        time.Time `json:"starts_at"`
	DurationMinutes int       `json:"duration_minutes"`
}

func (s *AppointmentService) CreateAppointmentForPatient(ctx context.Context, req *CreateAppointmentByPatientRequest) (*models.Appointment, error) {
	patient, err := s.patientRepo.FindByID(ctx, req.PatientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	appointment := &models.Appointment{
		PatientID:       req.PatientID,
		ProfessionalID:  req.ProfessionalID,
		StartsAt:        req.StartsAt,
		DurationMinutes: req.DurationMinutes,
		Status:          "scheduled",
	}
	if err := s.appointmentRepo.Create(ctx, appointment); err != nil {
		return nil, fmt.Errorf("failed to create appointment: %w", err)
	}
	appointment.Patient = patient

	message := fmt.Sprintf(
		"✅ Turno confirmado\n\nFecha: %s\nDuración: %d minutos\n\nGracias por reservar.",
		appointment.StartsAt.Format("02/01/2006 15:04"),
		appointment.DurationMinutes,
	)
	if patient.Email != "" && s.emailService != nil {
		s.emailService.SendAppointmentConfirmation(ctx, patient.Email, patient.Name, appointment.StartsAt, appointment.DurationMinutes)
	} else if patient.Phone != "" {
		if err := s.whatsappSender.SendMessage(ctx, patient.Phone, message); err != nil {
			fmt.Printf("Warning: failed to send WhatsApp confirmation: %v\n", err)
		}
	}

	return appointment, nil
}

func (s *AppointmentService) GetAppointment(ctx context.Context, id int64) (*models.Appointment, error) {
	return s.appointmentRepo.FindByID(ctx, id)
}

func (s *AppointmentService) GetAllAppointments(ctx context.Context) ([]*models.Appointment, error) {
	return s.appointmentRepo.FindAll(ctx)
}

type UpdateAppointmentRequest struct {
	Status          string     `json:"status"`
	Notes           string     `json:"notes"`
	StartsAt        *time.Time `json:"starts_at"`
	DurationMinutes *int       `json:"duration_minutes"`
}

func (s *AppointmentService) UpdateAppointment(ctx context.Context, id int64, req *UpdateAppointmentRequest) error {
	appointment, err := s.appointmentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Status != "" {
		appointment.Status = req.Status
	}
	if req.Notes != "" {
		appointment.Notes = req.Notes
	}
	if req.StartsAt != nil {
		appointment.StartsAt = *req.StartsAt
	}
	if req.DurationMinutes != nil {
		appointment.DurationMinutes = *req.DurationMinutes
	}

	return s.appointmentRepo.Update(ctx, appointment)
}

func (s *AppointmentService) CancelAppointment(ctx context.Context, id int64) error {
	return s.appointmentRepo.Delete(ctx, id)
}

// SendReminders busca turnos que necesitan recordatorio y envía WhatsApp
func (s *AppointmentService) SendReminders(ctx context.Context) error {
	// Buscar turnos entre 23-25 horas desde ahora que no tienen recordatorio enviado
	now := time.Now()
	from := now.Add(23 * time.Hour)
	to := now.Add(25 * time.Hour)

	appointments, err := s.appointmentRepo.FindPendingReminders(ctx, from, to)
	if err != nil {
		return fmt.Errorf("failed to find pending reminders: %w", err)
	}

	for _, appointment := range appointments {
		// Cargar paciente si no está cargado
		if appointment.Patient == nil {
			patient, err := s.patientRepo.FindByID(ctx, appointment.PatientID)
			if err != nil {
				fmt.Printf("Warning: failed to load patient %d: %v\n", appointment.PatientID, err)
				continue
			}
			appointment.Patient = patient
		}

		// Enviar recordatorio por email o WhatsApp
		if appointment.Patient.Email != "" && s.emailService != nil {
			s.emailService.SendAppointmentReminder(ctx, appointment.Patient.Email, appointment.Patient.Name, appointment.StartsAt, appointment.DurationMinutes)
		} else if appointment.Patient.Phone != "" {
			message := fmt.Sprintf(
				"📅 Recordatorio de turno\n\nTienes un turno mañana a las %s\nDuración: %d minutos\n\n¡Te esperamos!",
				appointment.StartsAt.Format("15:04"),
				appointment.DurationMinutes,
			)
			if err := s.whatsappSender.SendMessage(ctx, appointment.Patient.Phone, message); err != nil {
				fmt.Printf("Warning: failed to send reminder for appointment %d: %v\n", appointment.ID, err)
				continue
			}
		}

		// Marcar recordatorio como enviado
		now := time.Now()
		appointment.ReminderSentAt = &now
		if err := s.appointmentRepo.Update(ctx, appointment); err != nil {
			fmt.Printf("Warning: failed to update reminder timestamp for appointment %d: %v\n", appointment.ID, err)
		}
	}

	return nil
}
