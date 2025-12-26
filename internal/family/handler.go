package family

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.listFamilies)
	rg.POST("", h.createFamily)
	rg.GET("/:familyId", h.getFamily)

	rg.POST("/:familyId/invite", h.inviteMember)
	rg.DELETE("/:familyId/members/:userId", h.removeMember)

	rg.GET("/:familyId/children", h.listChildren)
	rg.POST("/:familyId/children", h.addChild)
	rg.PUT("/:familyId/children/:childId", h.updateChild)
	rg.DELETE("/:familyId/children/:childId", h.deleteChild)
}

func (h *Handler) listFamilies(c *gin.Context) {
	userID := c.GetString("user_id") // from auth middleware
	families, err := h.service.GetUserFamilies(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, families)
}

func (h *Handler) createFamily(c *gin.Context) {
	var req CreateFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetString("user_id")
	family, err := h.service.CreateFamily(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, family)
}

func (h *Handler) getFamily(c *gin.Context) {
	familyID := c.Param("familyId")
	family, err := h.service.GetFamily(c.Request.Context(), familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, family)
}

func (h *Handler) inviteMember(c *gin.Context) {
	var req InviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	familyID := c.Param("familyId")
	if err := h.service.InviteMember(c.Request.Context(), familyID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "invite sent"})
}

func (h *Handler) removeMember(c *gin.Context) {
	familyID := c.Param("familyId")
	userID := c.Param("userId")
	if err := h.service.RemoveMember(c.Request.Context(), familyID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) listChildren(c *gin.Context) {
	familyID := c.Param("familyId")
	children, err := h.service.GetChildren(c.Request.Context(), familyID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, children)
}

func (h *Handler) addChild(c *gin.Context) {
	var req AddChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	familyID := c.Param("familyId")
	child, err := h.service.AddChild(c.Request.Context(), familyID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, child)
}

func (h *Handler) updateChild(c *gin.Context) {
	var req AddChildRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	childID := c.Param("childId")
	child, err := h.service.UpdateChild(c.Request.Context(), childID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, child)
}

func (h *Handler) deleteChild(c *gin.Context) {
	childID := c.Param("childId")
	if err := h.service.DeleteChild(c.Request.Context(), childID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
