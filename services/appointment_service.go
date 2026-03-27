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
	FindAll(ctx context.Context, professionalID int64) ([]*models.Appointment, error)
	FindByPatient(ctx context.Context, patientID int64) ([]*models.Appointment, error)
	Update(ctx context.Context, appointment *models.Appointment) error
	Delete(ctx context.Context, id int64) error
	FindByDate(ctx context.Context, professionalID int64, from, to time.Time) ([]*models.Appointment, error)
	FindPendingReminders(ctx context.Context, from, to time.Time) ([]*models.Appointment, error)
}

// PatientRepository define métodos para acceso a pacientes
type PatientRepository interface {
	Create(ctx context.Context, patient *models.Patient) error
	Update(ctx context.Context, patient *models.Patient) error
	FindByPhone(ctx context.Context, phone string, professionalID int64) (*models.Patient, error)
	FindByID(ctx context.Context, id int64) (*models.Patient, error)
	FindAll(ctx context.Context, professionalID int64) ([]*models.Patient, error)
}

// ProfessionalLookup permite obtener datos del profesional
type ProfessionalLookup interface {
	FindByID(ctx context.Context, id int64) (*models.Professional, error)
}

// AppointmentService maneja lógica de turnos
type AppointmentService struct {
	appointmentRepo  AppointmentRepository
	patientRepo      PatientRepository
	professionalRepo ProfessionalLookup
	whatsappSender   WhatsAppSender
	emailService     *EmailService
}

func NewAppointmentService(
	appointmentRepo AppointmentRepository,
	patientRepo PatientRepository,
	professionalRepo ProfessionalLookup,
	whatsappSender WhatsAppSender,
	emailService *EmailService,
) *AppointmentService {
	return &AppointmentService{
		appointmentRepo:  appointmentRepo,
		patientRepo:      patientRepo,
		professionalRepo: professionalRepo,
		whatsappSender:   whatsappSender,
		emailService:     emailService,
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
	patient, err := s.patientRepo.FindByPhone(ctx, req.PatientPhone, req.ProfessionalID)
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

	// Enviar confirmación por WhatsApp al paciente
	message := fmt.Sprintf(
		"✅ Turno confirmado\n\nFecha: %s\nDuración: %d minutos\n\nGracias por reservar.",
		appointment.StartsAt.Format("02/01/2006 15:04"),
		appointment.DurationMinutes,
	)
	if err := s.whatsappSender.SendMessage(ctx, patient.Phone, message); err != nil {
		fmt.Printf("Warning: failed to send WhatsApp confirmation: %v\n", err)
	}

	// Notificar al profesional por email
	if s.emailService != nil {
		if prof, err := s.professionalRepo.FindByID(ctx, req.ProfessionalID); err == nil && prof.Email != "" {
			s.emailService.SendNewBookingNotification(ctx, prof.Email, prof.Name, patient.Name, appointment.StartsAt, appointment.DurationMinutes)
		}
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

type CreateRecurringRequest struct {
	PatientID       int64     `json:"patient_id"`
	ProfessionalID  int64     `json:"professional_id"`
	StartsAt        time.Time `json:"starts_at"`
	DurationMinutes int       `json:"duration_minutes"`
	FrequencyWeeks  int       `json:"frequency_weeks"` // 1 = semanal, 2 = quincenal
	Occurrences     int       `json:"occurrences"`     // cantidad de turnos a crear
}

func (s *AppointmentService) CreateRecurringAppointments(ctx context.Context, req *CreateRecurringRequest) ([]*models.Appointment, error) {
	patient, err := s.patientRepo.FindByID(ctx, req.PatientID)
	if err != nil {
		return nil, fmt.Errorf("patient not found: %w", err)
	}

	if req.FrequencyWeeks < 1 {
		req.FrequencyWeeks = 1
	}
	if req.Occurrences < 1 || req.Occurrences > 52 {
		req.Occurrences = 1
	}

	var created []*models.Appointment
	for i := 0; i < req.Occurrences; i++ {
		startsAt := req.StartsAt.AddDate(0, 0, i*req.FrequencyWeeks*7)
		appointment := &models.Appointment{
			PatientID:       req.PatientID,
			ProfessionalID:  req.ProfessionalID,
			StartsAt:        startsAt,
			DurationMinutes: req.DurationMinutes,
			Status:          "scheduled",
		}
		if err := s.appointmentRepo.Create(ctx, appointment); err != nil {
			return nil, fmt.Errorf("failed to create appointment %d: %w", i+1, err)
		}
		appointment.Patient = patient
		created = append(created, appointment)
	}

	// Notificar solo el primer turno
	if len(created) > 0 && patient.Email != "" && s.emailService != nil {
		s.emailService.SendAppointmentConfirmation(ctx, patient.Email, patient.Name, created[0].StartsAt, req.DurationMinutes)
	}

	return created, nil
}

func (s *AppointmentService) GetAppointment(ctx context.Context, id int64) (*models.Appointment, error) {
	return s.appointmentRepo.FindByID(ctx, id)
}

func (s *AppointmentService) GetAllAppointments(ctx context.Context, professionalID int64) ([]*models.Appointment, error) {
	return s.appointmentRepo.FindAll(ctx, professionalID)
}

func (s *AppointmentService) GetAppointmentsByPatient(ctx context.Context, patientID int64) ([]*models.Appointment, error) {
	return s.appointmentRepo.FindByPatient(ctx, patientID)
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
