package main

import (
	"fmt"
	"log"

	"github.com/andresramirez/psych-appointments/config"
	"github.com/andresramirez/psych-appointments/controllers"
	"github.com/andresramirez/psych-appointments/db"
	"github.com/andresramirez/psych-appointments/repositories"
	"github.com/andresramirez/psych-appointments/router"
	"github.com/andresramirez/psych-appointments/scheduler"
	"github.com/andresramirez/psych-appointments/services"
)

func main() {
	// 1. Cargar configuración
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("✅ Configuration loaded")

	// 2. Conectar a base de datos
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	log.Println("✅ Database connected and migrated")

	// 3. Inicializar repositorios
	professionalRepo := repositories.NewProfessionalRepository(database)
	patientRepo := repositories.NewPatientRepository(database)
	appointmentRepo := repositories.NewAppointmentRepository(database)
	availabilityRepo := repositories.NewAvailabilityRepository(database)
	noteRepo := repositories.NewNoteRepository(database)
	blockRepo := repositories.NewBlockRepository(database)

	log.Println("✅ Repositories initialized")

	// 4. Inicializar WhatsApp sender (mock o evolution)
	var whatsappSender services.WhatsAppSender
	if cfg.WhatsApp.Mode == "evolution" {
		log.Println("📱 WhatsApp mode: Evolution API")
		whatsappSender = services.NewEvolutionWhatsAppClient(
			cfg.WhatsApp.APIURL,
			cfg.WhatsApp.APIKey,
			cfg.WhatsApp.InstanceName,
		)
	} else {
		log.Println("📱 WhatsApp mode: Mock (development)")
		whatsappSender = services.NewMockWhatsAppSender()
	}

	// 5. Inicializar servicios
	authService := services.NewAuthService(professionalRepo, cfg.JWTSecret)

	var emailService *services.EmailService
	if cfg.Email.ResendAPIKey != "" {
		emailService = services.NewEmailService(cfg.Email.ResendAPIKey, cfg.Email.FromEmail)
		log.Println("📧 Email service: Resend")
	} else {
		log.Println("📧 Email service: disabled (RESEND_API_KEY not set)")
	}

	appointmentService := services.NewAppointmentService(appointmentRepo, patientRepo, whatsappSender, emailService)
	availabilityService := services.NewAvailabilityService(availabilityRepo, appointmentRepo)
	patientService := services.NewPatientService(patientRepo)
	noteService := services.NewNoteService(noteRepo)
	blockService := services.NewBlockService(blockRepo)

	log.Println("✅ Services initialized")

	// 6. Inicializar controladores
	authController := controllers.NewAuthController(authService)
	appointmentController := controllers.NewAppointmentController(appointmentService)
	availabilityController := controllers.NewAvailabilityController(availabilityService)
	patientController := controllers.NewPatientController(patientService)
	noteController := controllers.NewNoteController(noteService)
	publicController := controllers.NewPublicController(availabilityService, appointmentService, authService)
	blockController := controllers.NewBlockController(blockService)

	log.Println("✅ Controllers initialized")

	// 7. Configurar router
	r := router.NewRouter(
		authService,
		authController,
		appointmentController,
		availabilityController,
		patientController,
		noteController,
		publicController,
		blockController,
	)

	log.Println("✅ Router configured")

	// 8. Iniciar scheduler de recordatorios
	reminderScheduler := scheduler.NewScheduler(appointmentService)
	reminderScheduler.Start()

	log.Println("✅ Reminder scheduler started")

	// 9. Iniciar servidor
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("🚀 Server starting on %s", addr)
	log.Println("========================================")
	log.Println("Ready to accept requests!")
	log.Println("========================================")

	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
