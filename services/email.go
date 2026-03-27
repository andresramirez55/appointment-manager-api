package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/resend/resend-go/v2"
)

type EmailService struct {
	client    *resend.Client
	fromEmail string
}

func NewEmailService(apiKey, fromEmail string) *EmailService {
	return &EmailService{
		client:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
	}
}

func (s *EmailService) SendAppointmentConfirmation(ctx context.Context, toEmail, patientName string, startsAt time.Time, durationMinutes int) {
	if toEmail == "" {
		return
	}

	dateStr := startsAt.Format("02/01/2006")
	timeStr := startsAt.Format("15:04")

	html := fmt.Sprintf(`
<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto; padding: 32px; color: #1e293b;">
  <h2 style="font-size: 18px; font-weight: 600; margin-bottom: 8px;">✅ Turno confirmado</h2>
  <p style="color: #64748b; margin-bottom: 24px;">Hola %s, tu turno fue confirmado.</p>
  <div style="background: #f8fafc; border-radius: 12px; padding: 16px 20px; margin-bottom: 24px;">
    <p style="margin: 0 0 8px; font-size: 14px;"><strong>Fecha:</strong> %s</p>
    <p style="margin: 0 0 8px; font-size: 14px;"><strong>Hora:</strong> %s</p>
    <p style="margin: 0; font-size: 14px;"><strong>Duración:</strong> %d minutos</p>
  </div>
  <p style="color: #94a3b8; font-size: 12px;">Si necesitás reprogramar, contactá al profesional.</p>
</div>`, patientName, dateStr, timeStr, durationMinutes)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Turno confirmado para el %s a las %s", dateStr, timeStr),
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("Warning: failed to send confirmation email to %s: %v", toEmail, err)
	}
}

func (s *EmailService) SendNewBookingNotification(ctx context.Context, toEmail, professionalName, patientName string, startsAt time.Time, durationMinutes int) {
	if toEmail == "" {
		return
	}

	dateStr := startsAt.Format("02/01/2006")
	timeStr := startsAt.Format("15:04")

	html := fmt.Sprintf(`
<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto; padding: 32px; color: #1e293b;">
  <h2 style="font-size: 18px; font-weight: 600; margin-bottom: 8px;">📋 Nueva reserva de turno</h2>
  <p style="color: #64748b; margin-bottom: 24px;">Hola %s, un paciente acaba de reservar un turno.</p>
  <div style="background: #f8fafc; border-radius: 12px; padding: 16px 20px; margin-bottom: 24px;">
    <p style="margin: 0 0 8px; font-size: 14px;"><strong>Paciente:</strong> %s</p>
    <p style="margin: 0 0 8px; font-size: 14px;"><strong>Fecha:</strong> %s</p>
    <p style="margin: 0 0 8px; font-size: 14px;"><strong>Hora:</strong> %s</p>
    <p style="margin: 0; font-size: 14px;"><strong>Duración:</strong> %d minutos</p>
  </div>
  <p style="color: #94a3b8; font-size: 12px;">Ingresá a la app para ver el detalle.</p>
</div>`, professionalName, patientName, dateStr, timeStr, durationMinutes)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Nueva reserva: %s el %s a las %s", patientName, dateStr, timeStr),
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("Warning: failed to send booking notification to professional %s: %v", toEmail, err)
	}
}

func (s *EmailService) SendAppointmentReminder(ctx context.Context, toEmail, patientName string, startsAt time.Time, durationMinutes int) {
	if toEmail == "" {
		return
	}

	timeStr := startsAt.Format("15:04")

	html := fmt.Sprintf(`
<div style="font-family: sans-serif; max-width: 480px; margin: 0 auto; padding: 32px; color: #1e293b;">
  <h2 style="font-size: 18px; font-weight: 600; margin-bottom: 8px;">📅 Recordatorio de turno</h2>
  <p style="color: #64748b; margin-bottom: 24px;">Hola %s, te recordamos que mañana tenés turno.</p>
  <div style="background: #f8fafc; border-radius: 12px; padding: 16px 20px; margin-bottom: 24px;">
    <p style="margin: 0 0 8px; font-size: 14px;"><strong>Hora:</strong> %s</p>
    <p style="margin: 0; font-size: 14px;"><strong>Duración:</strong> %d minutos</p>
  </div>
  <p style="color: #94a3b8; font-size: 12px;">¡Te esperamos!</p>
</div>`, patientName, timeStr, durationMinutes)

	params := &resend.SendEmailRequest{
		From:    s.fromEmail,
		To:      []string{toEmail},
		Subject: "Recordatorio: tenés turno mañana",
		Html:    html,
	}

	_, err := s.client.Emails.Send(params)
	if err != nil {
		log.Printf("Warning: failed to send reminder email to %s: %v", toEmail, err)
	}
}
