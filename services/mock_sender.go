package services

import (
	"context"
	"log"
)

// MockWhatsAppSender implementa WhatsAppSender para desarrollo sin Evolution API
type MockWhatsAppSender struct{}

func NewMockWhatsAppSender() WhatsAppSender {
	return &MockWhatsAppSender{}
}

func (m *MockWhatsAppSender) SendMessage(ctx context.Context, phone string, message string) error {
	log.Printf("[MOCK WhatsApp] To: %s | Message: %s", phone, message)
	return nil
}
