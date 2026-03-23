package db

import (
	"fmt"
	"log"

	"github.com/andresramirez/psych-appointments/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect establece la conexión con la base de datos y ejecuta migraciones
func Connect(databaseURL string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Ejecutar migraciones automáticas
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed de datos iniciales
	if err := seedData(db); err != nil {
		return nil, fmt.Errorf("failed to seed data: %w", err)
	}

	return db, nil
}

// runMigrations ejecuta las migraciones de GORM
func runMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	return db.AutoMigrate(
		&models.Professional{},
		&models.Patient{},
		&models.AvailabilitySlot{},
		&models.AvailabilityOverride{},
		&models.Appointment{},
		&models.SessionNote{},
		&models.Block{},
	)
}

// seedData crea datos iniciales si no existen
func seedData(db *gorm.DB) error {
	log.Println("Seeding initial data...")

	// Verificar si ya existe un profesional
	var count int64
	if err := db.Model(&models.Professional{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		log.Println("Professional already exists, skipping seed")
		return nil
	}

	// Crear profesional por defecto
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	professional := &models.Professional{
		Email:     "admin@test.com",
		Password:  string(hashedPassword),
		Name:      "Dr. Admin",
		Phone:     "+5491112345678",
		Specialty: "Psicología",
	}

	if err := db.Create(professional).Error; err != nil {
		return fmt.Errorf("failed to create default professional: %w", err)
	}

	log.Printf("✅ Created default professional: %s (password: admin123)", professional.Email)
	return nil
}
