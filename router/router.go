package router

import (
	"net/http"

	"github.com/andresramirez/psych-appointments/controllers"
	"github.com/andresramirez/psych-appointments/middleware"
	"github.com/andresramirez/psych-appointments/services"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(
	authService *services.AuthService,
	authController *controllers.AuthController,
	appointmentController *controllers.AppointmentController,
	availabilityController *controllers.AvailabilityController,
	patientController *controllers.PatientController,
	noteController *controllers.NoteController,
	publicController *controllers.PublicController,
) *Router {
	engine := gin.Default()

	engine.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	// Health check
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := engine.Group("/api")

	// Public routes (sin autenticación)
	public := api.Group("/public")
	{
		public.GET("/slots", publicController.GetAvailableSlots)
		public.POST("/appointments", publicController.CreateAppointment)
	}

	// Auth
	api.POST("/auth/login", authController.Login)

	// Protected routes (requieren autenticación)
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		// Appointments
		appointments := protected.Group("/appointments")
		{
			appointments.POST("", appointmentController.Create)
			appointments.POST("/recurring", appointmentController.CreateRecurring)
			appointments.GET("", appointmentController.GetAll)
			appointments.GET("/:id", appointmentController.GetByID)
			appointments.PUT("/:id", appointmentController.Update)
			appointments.DELETE("/:id", appointmentController.Delete)
		}

		// Availability
		availability := protected.Group("/availability")
		{
			availability.POST("", availabilityController.CreateSlot)
			availability.GET("", availabilityController.GetSlots)
			availability.DELETE("/:id", availabilityController.DeleteSlot)
			availability.POST("/overrides", availabilityController.CreateOverride)
		}

		// Patients
		patients := protected.Group("/patients")
		{
			patients.POST("", patientController.Create)
			patients.GET("", patientController.GetAll)
			patients.GET("/:id", patientController.GetByID)
			patients.PUT("/:id", patientController.Update)
		}

		// Notes
		notes := protected.Group("/notes")
		{
			notes.POST("", noteController.Create)
			notes.GET("", noteController.GetByAppointment)
		}
	}

	return &Router{engine: engine}
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
