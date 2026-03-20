package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"

	"github.com/adedejiosvaldo/safetrace/backend/internal/config"
	"github.com/adedejiosvaldo/safetrace/backend/internal/database"
	"github.com/adedejiosvaldo/safetrace/backend/internal/handlers"
	"github.com/adedejiosvaldo/safetrace/backend/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Postgres
	postgres, err := database.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to Postgres: %v", err)
	}
	defer postgres.Close()
	log.Println("✓ Connected to Postgres")

	// Initialize Redis
	redis, err := database.NewRedisDB(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()
	log.Println("✓ Connected to Redis")

	// Initialize Firebase (optional)
	var fcmClient *messaging.Client
	if cfg.FCMCredentialsPath != "" {
		ctx := context.Background()
		opt := option.WithCredentialsFile(cfg.FCMCredentialsPath)
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			log.Printf("Warning: Failed to initialize Firebase: %v", err)
		} else {
			fcmClient, err = app.Messaging(ctx)
			if err != nil {
				log.Printf("Warning: Failed to initialize FCM client: %v", err)
			} else {
				log.Println("✓ Firebase FCM initialized")
			}
		}
	}

	// Initialize services
	alertEngine := services.NewAlertEngine(cfg, fcmClient)
	evaluator := services.NewSafetyEvaluator(cfg, postgres, redis, alertEngine)
	
	tripService := services.NewTripService(postgres)
	timerService := services.NewTimerService(postgres, alertEngine)
	incidentService := services.NewIncidentService(postgres)
	settingsService := services.NewSettingsService(postgres)
	
	log.Println("✓ Services initialized")

	// Initialize handlers
	heartbeatHandler := handlers.NewHeartbeatHandler(cfg, postgres, redis, evaluator, alertEngine)
	smsHandler := handlers.NewSMSHandler(cfg, postgres, redis, evaluator)
	blackboxHandler := handlers.NewBlackboxHandler(cfg, postgres)
	contactsHandler := handlers.NewContactsHandler(cfg, postgres)
	
	tripHandler := handlers.NewTripHandler(tripService)
	timerHandler := handlers.NewTimerHandler(timerService)
	incidentHandler := handlers.NewIncidentHandler(incidentService)
	settingsHandler := handlers.NewSettingsHandler(settingsService)

	// Setup Gin router
	router := setupRouter(
		heartbeatHandler, smsHandler, blackboxHandler, contactsHandler,
		tripHandler, timerHandler, incidentHandler, settingsHandler,
	)

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Printf("🚀 SafeTrace API server starting on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped gracefully")
}

func setupRouter(
	heartbeatHandler *handlers.HeartbeatHandler,
	smsHandler *handlers.SMSHandler,
	blackboxHandler *handlers.BlackboxHandler,
	contactsHandler *handlers.ContactsHandler,
	tripHandler *handlers.TripHandler,
	timerHandler *handlers.TimerHandler,
	incidentHandler *handlers.IncidentHandler,
	settingsHandler *handlers.SettingsHandler,
) *gin.Engine {
	router := gin.Default()

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "safetrace-api",
			"time":    time.Now().Format(time.RFC3339),
		})
	})

	// API v1 routes
	v1 := router.Group("/v1")
	{
		// Heartbeat & Alerts endpoints
		v1.POST("/heartbeat", heartbeatHandler.CreateHeartbeat)
		v1.GET("/user/:id/status", heartbeatHandler.GetUserStatus)
		v1.POST("/alert/trigger", heartbeatHandler.TriggerAlert)
		v1.POST("/alert/:id/resolve", heartbeatHandler.ResolveAlert)

		// SMS webhook
		v1.POST("/sms/webhook", smsHandler.HandleIncomingSMS)

		// Blackbox endpoints
		v1.POST("/blackbox/upload", blackboxHandler.UploadTrail)
		v1.GET("/blackbox/trails/:user_id", blackboxHandler.GetUserTrails)

		// Contact management endpoints
		v1.GET("/users/search", contactsHandler.SearchUsers)
		v1.GET("/user/:id/contacts", contactsHandler.GetContacts)
		v1.POST("/user/:id/contacts", contactsHandler.AddContact)
		v1.PUT("/user/:id/contacts/:contactId", contactsHandler.UpdateContact)
		v1.DELETE("/user/:id/contacts/:contactId", contactsHandler.DeleteContact)

		// Live Trip endpoints
		v1.POST("/trips/start", tripHandler.StartTrip)
		v1.POST("/trips/:id/location", tripHandler.StreamLocation)
		v1.POST("/trips/:id/end", tripHandler.EndTrip)
		v1.GET("/trips/:id/guardians", tripHandler.GetGuardians)

		// Timer endpoints
		v1.POST("/timers/start", timerHandler.StartTimer)
		v1.POST("/timers/:id/extend", timerHandler.ExtendTimer)
		v1.POST("/timers/:id/safe", timerHandler.MarkSafe)

		// Incidents
		v1.POST("/incidents/report", incidentHandler.ReportIncident)

		// Settings & Data
		v1.PUT("/user/:id/settings", settingsHandler.UpdateSettings)
		v1.DELETE("/user/:id/data", settingsHandler.DeleteData)
	}

	return router
}
