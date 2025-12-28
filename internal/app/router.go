package app

import "github.com/gin-gonic/gin"

func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Version endpoint
		api.GET("/version", func(c *gin.Context) {
			c.JSON(200, gin.H{"version": GetVersion()})
		})

		// Auth routes (public)
		authGroup := api.Group("/auth")
		s.authHandler.RegisterRoutes(authGroup)

		// Protected routes
		protected := api.Group("/")
		protected.Use(s.authMiddleware())
		{
			// Family routes
			familyGroup := protected.Group("/families")
			s.familyHandler.RegisterRoutes(familyGroup)

			// Feeding routes
			feedingGroup := protected.Group("/feeding")
			s.feedingHandler.RegisterRoutes(feedingGroup)

			// Sleep routes
			sleepGroup := protected.Group("/sleep")
			s.sleepHandler.RegisterRoutes(sleepGroup)

			// Medication routes
			medicationGroup := protected.Group("/medications")
			s.medicationHandler.RegisterRoutes(medicationGroup)

			// Vaccination routes
			vaccinationGroup := protected.Group("/vaccinations")
			s.vaccinationHandler.RegisterRoutes(vaccinationGroup)

			// Appointment routes
			appointmentGroup := protected.Group("/appointments")
			s.appointmentHandler.RegisterRoutes(appointmentGroup)

			// Notes routes
			notesGroup := protected.Group("/notes")
			s.notesHandler.RegisterRoutes(notesGroup)

			// Sync routes
			syncGroup := protected.Group("/sync")
			s.syncHandler.RegisterRoutes(syncGroup)

			// Notifications routes (SSE)
			notificationsGroup := protected.Group("/notifications")
			s.notificationsHandler.RegisterRoutes(notificationsGroup)
		}
	}

	// Serve UI for all other routes
	s.serveUI()
}
