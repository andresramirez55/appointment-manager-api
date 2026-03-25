package services

import (
	"context"
	"fmt"
	"time"

	"github.com/andresramirez/psych-appointments/models"
)

// AvailabilityRepository define métodos para acceso a disponibilidad
type AvailabilityRepository interface {
	CreateSlot(ctx context.Context, slot *models.AvailabilitySlot) error
	FindSlotsByProfessional(ctx context.Context, professionalID int64) ([]*models.AvailabilitySlot, error)
	UpdateSlot(ctx context.Context, slot *models.AvailabilitySlot) error
	DeleteSlot(ctx context.Context, id int64) error

	CreateOverride(ctx context.Context, override *models.AvailabilityOverride) error
	FindOverridesByProfessional(ctx context.Context, professionalID int64) ([]*models.AvailabilityOverride, error)
	FindOverrideByDate(ctx context.Context, professionalID int64, date time.Time) (*models.AvailabilityOverride, error)
}

// AvailabilityService maneja lógica de disponibilidad
type AvailabilityService struct {
	availabilityRepo AvailabilityRepository
	appointmentRepo  AppointmentRepository
}

func NewAvailabilityService(
	availabilityRepo AvailabilityRepository,
	appointmentRepo AppointmentRepository,
) *AvailabilityService {
	return &AvailabilityService{
		availabilityRepo: availabilityRepo,
		appointmentRepo:  appointmentRepo,
	}
}

type CreateSlotRequest struct {
	ProfessionalID      int64  `json:"professional_id"`
	DayOfWeek           int    `json:"day_of_week"`
	StartTime           string `json:"start_time"`
	EndTime             string `json:"end_time"`
	SlotDurationMinutes int    `json:"slot_duration_minutes"`
}

func (s *AvailabilityService) CreateSlot(ctx context.Context, req *CreateSlotRequest) (*models.AvailabilitySlot, error) {
	slot := &models.AvailabilitySlot{
		ProfessionalID:      req.ProfessionalID,
		DayOfWeek:           req.DayOfWeek,
		StartTime:           req.StartTime,
		EndTime:             req.EndTime,
		SlotDurationMinutes: req.SlotDurationMinutes,
	}

	if err := s.availabilityRepo.CreateSlot(ctx, slot); err != nil {
		return nil, fmt.Errorf("failed to create slot: %w", err)
	}

	return slot, nil
}

func (s *AvailabilityService) GetSlots(ctx context.Context, professionalID int64) ([]*models.AvailabilitySlot, error) {
	return s.availabilityRepo.FindSlotsByProfessional(ctx, professionalID)
}

func (s *AvailabilityService) UpdateSlot(ctx context.Context, slot *models.AvailabilitySlot) error {
	return s.availabilityRepo.UpdateSlot(ctx, slot)
}

func (s *AvailabilityService) DeleteSlot(ctx context.Context, id int64) error {
	return s.availabilityRepo.DeleteSlot(ctx, id)
}

type CreateOverrideRequest struct {
	ProfessionalID int64     `json:"professional_id"`
	Date           time.Time `json:"date"`
	Available      bool      `json:"available"`
	Reason         string    `json:"reason"`
	StartTime      string    `json:"start_time"`
	EndTime        string    `json:"end_time"`
}

func (s *AvailabilityService) CreateOverride(ctx context.Context, req *CreateOverrideRequest) (*models.AvailabilityOverride, error) {
	override := &models.AvailabilityOverride{
		ProfessionalID: req.ProfessionalID,
		Date:           req.Date,
		Available:      req.Available,
		Reason:         req.Reason,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
	}

	if err := s.availabilityRepo.CreateOverride(ctx, override); err != nil {
		return nil, fmt.Errorf("failed to create override: %w", err)
	}

	return override, nil
}

type TimeSlot struct {
	StartsAt time.Time `json:"starts_at"`
	EndsAt   time.Time `json:"ends_at"`
	Available bool     `json:"available"`
}

// GetAvailableSlots calcula slots disponibles para una fecha específica
func (s *AvailabilityService) GetAvailableSlots(ctx context.Context, professionalID int64, date time.Time) ([]*TimeSlot, error) {
	// Verificar si hay override para esta fecha
	override, err := s.availabilityRepo.FindOverrideByDate(ctx, professionalID, date)
	if err == nil && override != nil && !override.Available {
		// Día no disponible
		return []*TimeSlot{}, nil
	}

	// Obtener configuración regular para el día de la semana
	slots, err := s.availabilityRepo.FindSlotsByProfessional(ctx, professionalID)
	if err != nil {
		return nil, err
	}

	dayOfWeek := int(date.Weekday())
	var availabilitySlot *models.AvailabilitySlot
	for _, slot := range slots {
		if slot.DayOfWeek == dayOfWeek {
			availabilitySlot = slot
			break
		}
	}

	if availabilitySlot == nil {
		// No hay disponibilidad configurada para este día
		return []*TimeSlot{}, nil
	}

	// Parsear horarios
	startTime, _ := time.Parse("15:04", availabilitySlot.StartTime)
	endTime, _ := time.Parse("15:04", availabilitySlot.EndTime)

	// Generar slots
	var timeSlots []*TimeSlot
	currentStart := time.Date(date.Year(), date.Month(), date.Day(), startTime.Hour(), startTime.Minute(), 0, 0, date.Location())
	endDateTime := time.Date(date.Year(), date.Month(), date.Day(), endTime.Hour(), endTime.Minute(), 0, 0, date.Location())

	for currentStart.Before(endDateTime) {
		slotEnd := currentStart.Add(time.Duration(availabilitySlot.SlotDurationMinutes) * time.Minute)
		if slotEnd.After(endDateTime) {
			break
		}

		timeSlots = append(timeSlots, &TimeSlot{
			StartsAt:  currentStart,
			EndsAt:    slotEnd,
			Available: true,
		})

		currentStart = slotEnd
	}

	// Marcar slots ocupados
	dayEnd := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 0, date.Location())
	bookedAppts, err := s.appointmentRepo.FindByDate(ctx, professionalID, currentStart.Truncate(24*time.Hour), dayEnd)
	if err == nil {
		for _, appt := range bookedAppts {
			for _, slot := range timeSlots {
				if appt.StartsAt.Equal(slot.StartsAt) {
					slot.Available = false
				}
			}
		}
	}

	// Filtrar solo los disponibles
	var available []*TimeSlot
	for _, slot := range timeSlots {
		if slot.Available {
			available = append(available, slot)
		}
	}

	return available, nil
}
