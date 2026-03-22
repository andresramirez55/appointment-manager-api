package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/andresramirez/psych-appointments/services"
)

type Scheduler struct {
	appointmentService *services.AppointmentService
	ticker             *time.Ticker
	done               chan bool
}

func NewScheduler(appointmentService *services.AppointmentService) *Scheduler {
	return &Scheduler{
		appointmentService: appointmentService,
		done:               make(chan bool),
	}
}

// Start inicia el scheduler que verifica recordatorios cada hora
func (s *Scheduler) Start() {
	log.Println("Starting reminder scheduler (runs every hour)")

	// Ejecutar inmediatamente al arrancar
	s.checkReminders()

	// Configurar ticker para ejecutar cada hora
	s.ticker = time.NewTicker(1 * time.Hour)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				s.checkReminders()
			case <-s.done:
				log.Println("Scheduler stopped")
				return
			}
		}
	}()
}

func (s *Scheduler) checkReminders() {
	log.Println("⏰ Checking pending reminders...")

	ctx := context.Background()
	if err := s.appointmentService.SendReminders(ctx); err != nil {
		log.Printf("❌ Error sending reminders: %v", err)
	} else {
		log.Println("✅ Reminders check completed")
	}
}

// Stop detiene el scheduler
func (s *Scheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	s.done <- true
}
