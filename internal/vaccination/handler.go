package vaccination

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.list)
	rg.POST("", h.create)
	rg.GET("/schedule", h.getSchedule)
	rg.GET("/upcoming/:childId", h.getUpcoming)
	rg.POST("/generate/:childId", h.generateSchedule)
	rg.GET("/:id", h.get)
	rg.PUT("/:id", h.update)
	rg.DELETE("/:id", h.delete)
	rg.POST("/:id/record", h.recordAdministration)
}

func (h *Handler) list(c *gin.Context) {
	completed := c.Query("completed")
	var completedPtr *bool
	if completed != "" {
		val := completed == "true"
		completedPtr = &val
	}

	filter := &VaccinationFilter{
		ChildID:      c.Query("child_id"),
		Completed:    completedPtr,
		UpcomingOnly: c.Query("upcoming_only") == "true",
	}
	vaxes, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vaxes)
}

func (h *Handler) create(c *gin.Context) {
	var req CreateVaccinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vax, err := h.service.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, vax)
}

func (h *Handler) get(c *gin.Context) {
	id := c.Param("id")
	vax, err := h.service.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vax)
}

func (h *Handler) update(c *gin.Context) {
	var req CreateVaccinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	vax, err := h.service.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vax)
}

func (h *Handler) delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) recordAdministration(c *gin.Context) {
	var req RecordVaccinationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id := c.Param("id")
	vax, err := h.service.RecordAdministration(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vax)
}

func (h *Handler) getUpcoming(c *gin.Context) {
	childID := c.Param("childId")
	days := 30 // default
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}

	vaxes, err := h.service.GetUpcoming(c.Request.Context(), childID, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vaxes)
}

func (h *Handler) getSchedule(c *gin.Context) {
	schedule := h.service.GetSchedule()
	c.JSON(http.StatusOK, schedule)
}

func (h *Handler) generateSchedule(c *gin.Context) {
	childID := c.Param("childId")
	var req struct {
		BirthDate string `json:"birth_date" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vaxes, err := h.service.GenerateScheduleForChild(c.Request.Context(), childID, req.BirthDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, vaxes)
}
