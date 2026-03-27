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

	// Migración de datos de consultorios
	if err := migrateConsultorios(db); err != nil {
		return nil, fmt.Errorf("failed to migrate consultorios: %w", err)
	}

	return db, nil
}

// runMigrations ejecuta las migraciones de GORM
func runMigrations(db *gorm.DB) error {
	log.Println("Running database migrations...")

	return db.AutoMigrate(
		&models.Consultorio{},
		&models.Professional{},
		&models.Patient{},
		&models.AvailabilitySlot{},
		&models.AvailabilityOverride{},
		&models.Appointment{},
		&models.SessionNote{},
		&models.Block{},
	)
}

// migrateConsultorios crea consultorios iniciales y asigna pacientes/turnos existentes
func migrateConsultorios(db *gorm.DB) error {
	log.Println("Running consultorio data migration...")

	// 1. Crear consultorio "Principal" para cada profesional que no tenga uno
	if err := db.Exec(`
		INSERT INTO consultorios (professional_id, name, created_at, updated_at)
		SELECT p.id, 'Principal', NOW(), NOW()
		FROM professionals p
		WHERE p.deleted_at IS NULL
		AND NOT EXISTS (
			SELECT 1 FROM consultorios c
			WHERE c.professional_id = p.id AND c.deleted_at IS NULL
		)
	`).Error; err != nil {
		return err
	}

	// 2. Asignar pacientes sin consultorio al primer consultorio de su profesional
	if err := db.Exec(`
		UPDATE patients SET consultorio_id = (
			SELECT id FROM consultorios
			WHERE professional_id = patients.professional_id
			AND deleted_at IS NULL
			ORDER BY id ASC LIMIT 1
		)
		WHERE consultorio_id IS NULL AND deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	// 3. Asignar turnos sin consultorio al consultorio del paciente
	if err := db.Exec(`
		UPDATE appointments SET consultorio_id = (
			SELECT consultorio_id FROM patients
			WHERE patients.id = appointments.patient_id
			AND patients.deleted_at IS NULL
		)
		WHERE consultorio_id IS NULL AND deleted_at IS NULL
	`).Error; err != nil {
		return err
	}

	log.Println("Consultorio migration completed")
	return nil
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
