package services

import "context"

// WhatsAppSender define la interfaz para enviar mensajes de WhatsApp
type WhatsAppSender interface {
	SendMessage(ctx context.Context, phone string, message string) error
}
