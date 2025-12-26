package app

import "github.com/gin-gonic/gin"

func (s *Server) setupRoutes() {
	api := s.router.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
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
			vaccination := protected.Group("/vaccinations")
			{
				_ = vaccination // TODO: wire vaccination handlers
			}

			// Appointment routes
			appointment := protected.Group("/appointments")
			{
				_ = appointment // TODO: wire appointment handlers
			}

			// Notes routes
			notes := protected.Group("/notes")
			{
				_ = notes // TODO: wire notes handlers
			}

			// Sync routes
			sync := protected.Group("/sync")
			{
				_ = sync // TODO: wire sync handlers
			}
		}
	}

	// Serve UI for all other routes
	s.serveUI()
}
