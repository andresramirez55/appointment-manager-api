package models

import (
	"time"

	"gorm.io/gorm"
)

// Professional representa al profesional de la salud (admin)
type Professional struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	Password  string         `gorm:"not null" json:"-"` // hash bcrypt, no exponer en JSON
	Name      string         `gorm:"not null" json:"name"`
	Phone     string         `gorm:"not null" json:"phone"`
	Specialty string         `json:"specialty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Patient representa a un paciente
type Patient struct {
	ID        int64          `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Phone     string         `gorm:"not null;index" json:"phone"` // Usado para WhatsApp
	Email     string         `json:"email"`                       // Opcional
	Notes     string         `gorm:"type:text" json:"notes"`      // Notas generales del paciente
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// AvailabilitySlot representa un horario regular de disponibilidad
// Ejemplo: todos los lunes de 9 a 17hs con slots de 60 minutos
type AvailabilitySlot struct {
	ID                 int64          `gorm:"primaryKey" json:"id"`
	ProfessionalID     int64          `gorm:"not null;index" json:"professional_id"`
	Professional       *Professional  `gorm:"foreignKey:ProfessionalID" json:"professional,omitempty"`
	DayOfWeek          int            `gorm:"not null" json:"day_of_week"` // 0=Domingo, 1=Lunes, ..., 6=Sábado
	StartTime          string         `gorm:"not null" json:"start_time"`  // Formato "15:04"
	EndTime            string         `gorm:"not null" json:"end_time"`    // Formato "15:04"
	SlotDurationMinutes int           `gorm:"not null" json:"slot_duration_minutes"`
	CreatedAt          time.Time      `json:"created_at"`
	UpdatedAt          time.Time      `json:"updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
}

// AvailabilityOverride representa excepciones a la disponibilidad regular
// Ejemplo: feriado (available=false) o día especial (available=true con horarios custom)
type AvailabilityOverride struct {
	ID             int64          `gorm:"primaryKey" json:"id"`
	ProfessionalID int64          `gorm:"not null;index" json:"professional_id"`
	Professional   *Professional  `gorm:"foreignKey:ProfessionalID" json:"professional,omitempty"`
	Date           time.Time      `gorm:"not null;index" json:"date"` // Fecha específica
	Available      bool           `gorm:"not null" json:"available"`  // false = día no disponible
	Reason         string         `json:"reason"`                     // Ej: "Feriado", "Vacaciones"
	StartTime      string         `json:"start_time"`                 // Opcional: si available=true y quiere horario custom
	EndTime        string         `json:"end_time"`                   // Opcional
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// Appointment representa un turno
type Appointment struct {
	ID              int64          `gorm:"primaryKey" json:"id"`
	PatientID       int64          `gorm:"not null;index" json:"patient_id"`
	Patient         *Patient       `gorm:"foreignKey:PatientID" json:"patient,omitempty"`
	ProfessionalID  int64          `gorm:"not null;index" json:"professional_id"`
	Professional    *Professional  `gorm:"foreignKey:ProfessionalID" json:"professional,omitempty"`
	StartsAt        time.Time      `gorm:"not null;index" json:"starts_at"`
	DurationMinutes int            `gorm:"not null" json:"duration_minutes"`
	Status          string         `gorm:"not null;default:'scheduled'" json:"status"` // scheduled, completed, cancelled
	ReminderSentAt  *time.Time     `json:"reminder_sent_at"`                           // NULL si no se envió
	Notes           string         `gorm:"type:text" json:"notes"`                     // Notas del turno
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
}

// Block representa un bloqueo de agenda
type Block struct {
	ID             int64          `gorm:"primaryKey" json:"id"`
	ProfessionalID int64          `gorm:"not null;index" json:"professional_id"`
	StartsAt       time.Time      `gorm:"not null;index" json:"starts_at"`
	EndsAt         time.Time      `gorm:"not null" json:"ends_at"`
	Reason         string         `json:"reason"`
	CreatedAt      time.Time      `json:"created_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

// SessionNote representa notas de una sesión
type SessionNote struct {
	ID            int64          `gorm:"primaryKey" json:"id"`
	AppointmentID int64          `gorm:"not null;index" json:"appointment_id"`
	Appointment   *Appointment   `gorm:"foreignKey:AppointmentID" json:"appointment,omitempty"`
	Content       string         `gorm:"type:text;not null" json:"content"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}
