package router

import (
	"net/http"
	"os"
	"strings"
	"time"

	appointmentcontroller "github.com/andresramirez/psych-appointments/controllers/appointment"
	authcontroller "github.com/andresramirez/psych-appointments/controllers/auth"
	availabilitycontroller "github.com/andresramirez/psych-appointments/controllers/availability"
	blockcontroller "github.com/andresramirez/psych-appointments/controllers/block"
	notecontroller "github.com/andresramirez/psych-appointments/controllers/note"
	officecontroller "github.com/andresramirez/psych-appointments/controllers/office"
	patientcontroller "github.com/andresramirez/psych-appointments/controllers/patient"
	publiccontroller "github.com/andresramirez/psych-appointments/controllers/public"
	"github.com/andresramirez/psych-appointments/middleware"
	"github.com/andresramirez/psych-appointments/services"
	"github.com/gin-gonic/gin"
)

type Router struct {
	engine *gin.Engine
}

func NewRouter(
	authService *services.AuthService,
	authController *authcontroller.Controller,
	appointmentController *appointmentcontroller.Controller,
	availabilityController *availabilitycontroller.Controller,
	patientController *patientcontroller.Controller,
	noteController *notecontroller.Controller,
	publicController *publiccontroller.Controller,
	blockController *blockcontroller.Controller,
	officeController *officecontroller.Controller,
) *Router {
	engine := gin.Default()

	engine.Use(func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowed := strings.TrimRight(os.Getenv("FRONTEND_URL"), "/")
		if origin != "" {
			if origin != allowed {
				c.AbortWithStatus(http.StatusForbidden)
				return
			}
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Session-Request")
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

	// Public routes (sin autenticación) — rate limited
	public := api.Group("/public")
	public.Use(middleware.RateLimit(20, time.Minute))
	{
		public.GET("/professional/:id", publicController.GetProfessional)
		public.GET("/slots", publicController.GetAvailableSlots)
		public.POST("/appointments", publicController.CreateAppointment)
		public.GET("/appointments/:token", publicController.GetAppointmentByToken)
		public.POST("/appointments/:token/cancel", publicController.CancelByToken)
	}

	// Auth
	auth := api.Group("/auth")
	auth.Use(middleware.RateLimit(30, time.Minute))
	auth.Use(func(c *gin.Context) { c.Header("Cache-Control", "no-store"); c.Next() })
	auth.POST("/register", authController.Register)
	auth.POST("/login", authController.Login)
	auth.POST("/refresh", authController.Refresh)
	auth.POST("/logout", authController.Logout)

	// Protected routes (requieren autenticación)
	protected := api.Group("")
	protected.Use(func(c *gin.Context) { c.Header("Cache-Control", "no-store"); c.Next() })
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

		// Profile
		protected.GET("/profile", authController.GetProfile)
		protected.PUT("/profile", authController.UpdateProfile)

		// Blocks
		blocks := protected.Group("/blocks")
		{
			blocks.POST("", blockController.Create)
			blocks.GET("", blockController.GetAll)
			blocks.DELETE("/:id", blockController.Delete)
		}

		// Consultorios
		consultorios := protected.Group("/consultorios")
		{
			consultorios.GET("", officeController.GetAll)
			consultorios.POST("", officeController.Create)
			consultorios.PUT("/:id", officeController.Update)
			consultorios.DELETE("/:id", officeController.Delete)
		}
	}

	return &Router{engine: engine}
}

func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
