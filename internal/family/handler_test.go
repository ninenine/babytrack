package family

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// mockService is a test double for family.Service
type mockService struct {
	getUserFamiliesFn  func(ctx context.Context, userID string) ([]FamilyWithChildren, error)
	createFamilyFn     func(ctx context.Context, userID string, req *CreateFamilyRequest) (*Family, error)
	getFamilyFn        func(ctx context.Context, familyID string) (*Family, error)
	updateFamilyFn     func(ctx context.Context, familyID string, req *CreateFamilyRequest) (*Family, error)
	deleteFamilyFn     func(ctx context.Context, familyID, userID string) error
	leaveFamilyFn      func(ctx context.Context, familyID, userID string) error
	getMemberRoleFn    func(ctx context.Context, familyID, userID string) (string, error)
	getFamilyMembersFn func(ctx context.Context, familyID string) ([]MemberWithUser, error)
	inviteMemberFn     func(ctx context.Context, familyID string, req *InviteRequest) error
	joinFamilyFn       func(ctx context.Context, familyID, userID string) (*Family, error)
	removeMemberFn     func(ctx context.Context, familyID, userID string) error
	addChildFn         func(ctx context.Context, familyID string, req *AddChildRequest) (*Child, error)
	getChildrenFn      func(ctx context.Context, familyID string) ([]Child, error)
	getChildFn         func(ctx context.Context, childID string) (*Child, error)
	updateChildFn      func(ctx context.Context, childID string, req *AddChildRequest) (*Child, error)
	deleteChildFn      func(ctx context.Context, childID string) error
}

func (m *mockService) GetUserFamilies(ctx context.Context, userID string) ([]FamilyWithChildren, error) {
	if m.getUserFamiliesFn != nil {
		return m.getUserFamiliesFn(ctx, userID)
	}
	return nil, nil
}

func (m *mockService) CreateFamily(ctx context.Context, userID string, req *CreateFamilyRequest) (*Family, error) {
	if m.createFamilyFn != nil {
		return m.createFamilyFn(ctx, userID, req)
	}
	return nil, nil
}

func (m *mockService) GetFamily(ctx context.Context, familyID string) (*Family, error) {
	if m.getFamilyFn != nil {
		return m.getFamilyFn(ctx, familyID)
	}
	return nil, nil
}

func (m *mockService) UpdateFamily(ctx context.Context, familyID string, req *CreateFamilyRequest) (*Family, error) {
	if m.updateFamilyFn != nil {
		return m.updateFamilyFn(ctx, familyID, req)
	}
	return nil, nil
}

func (m *mockService) DeleteFamily(ctx context.Context, familyID, userID string) error {
	if m.deleteFamilyFn != nil {
		return m.deleteFamilyFn(ctx, familyID, userID)
	}
	return nil
}

func (m *mockService) LeaveFamily(ctx context.Context, familyID, userID string) error {
	if m.leaveFamilyFn != nil {
		return m.leaveFamilyFn(ctx, familyID, userID)
	}
	return nil
}

func (m *mockService) GetMemberRole(ctx context.Context, familyID, userID string) (string, error) {
	if m.getMemberRoleFn != nil {
		return m.getMemberRoleFn(ctx, familyID, userID)
	}
	return "", nil
}

func (m *mockService) GetFamilyMembers(ctx context.Context, familyID string) ([]MemberWithUser, error) {
	if m.getFamilyMembersFn != nil {
		return m.getFamilyMembersFn(ctx, familyID)
	}
	return nil, nil
}

func (m *mockService) InviteMember(ctx context.Context, familyID string, req *InviteRequest) error {
	if m.inviteMemberFn != nil {
		return m.inviteMemberFn(ctx, familyID, req)
	}
	return nil
}

func (m *mockService) JoinFamily(ctx context.Context, familyID, userID string) (*Family, error) {
	if m.joinFamilyFn != nil {
		return m.joinFamilyFn(ctx, familyID, userID)
	}
	return nil, nil
}

func (m *mockService) RemoveMember(ctx context.Context, familyID, userID string) error {
	if m.removeMemberFn != nil {
		return m.removeMemberFn(ctx, familyID, userID)
	}
	return nil
}

func (m *mockService) AddChild(ctx context.Context, familyID string, req *AddChildRequest) (*Child, error) {
	if m.addChildFn != nil {
		return m.addChildFn(ctx, familyID, req)
	}
	return nil, nil
}

func (m *mockService) GetChildren(ctx context.Context, familyID string) ([]Child, error) {
	if m.getChildrenFn != nil {
		return m.getChildrenFn(ctx, familyID)
	}
	return nil, nil
}

func (m *mockService) GetChild(ctx context.Context, childID string) (*Child, error) {
	if m.getChildFn != nil {
		return m.getChildFn(ctx, childID)
	}
	return nil, nil
}

func (m *mockService) UpdateChild(ctx context.Context, childID string, req *AddChildRequest) (*Child, error) {
	if m.updateChildFn != nil {
		return m.updateChildFn(ctx, childID, req)
	}
	return nil, nil
}

func (m *mockService) DeleteChild(ctx context.Context, childID string) error {
	if m.deleteChildFn != nil {
		return m.deleteChildFn(ctx, childID)
	}
	return nil
}

// setupRouter creates a test router with auth context middleware
func setupRouter(h *Handler) *gin.Engine {
	router := gin.New()
	// Middleware to set user_id in context (simulating auth middleware)
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "test-user")
		c.Next()
	})
	families := router.Group("/families")
	h.RegisterRoutes(families)
	return router
}

// setupRouterWithUserID creates a test router with a specific user_id
func setupRouterWithUserID(h *Handler, userID string) *gin.Engine {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	families := router.Group("/families")
	h.RegisterRoutes(families)
	return router
}

// ============================================================================
// List Families Tests
// ============================================================================

func TestListFamilies_Success(t *testing.T) {
	now := time.Now()
	mock := &mockService{
		getUserFamiliesFn: func(ctx context.Context, userID string) ([]FamilyWithChildren, error) {
			if userID != "test-user" {
				t.Errorf("Expected userID test-user, got %s", userID)
			}
			return []FamilyWithChildren{
				{
					ID:        "family-1",
					Name:      "Smith Family",
					Children:  []Child{{ID: "child-1", Name: "Alice"}},
					CreatedAt: now,
					UpdatedAt: now,
				},
				{
					ID:        "family-2",
					Name:      "Jones Family",
					Children:  []Child{},
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("GET", "/families", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []FamilyWithChildren
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 families, got %d", len(response))
	}

	if response[0].Name != "Smith Family" {
		t.Errorf("Expected Smith Family, got %s", response[0].Name)
	}
}

func TestListFamilies_Empty(t *testing.T) {
	mock := &mockService{
		getUserFamiliesFn: func(ctx context.Context, userID string) ([]FamilyWithChildren, error) {
			return []FamilyWithChildren{}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("GET", "/families", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []FamilyWithChildren
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response) != 0 {
		t.Errorf("Expected 0 families, got %d", len(response))
	}
}

func TestListFamilies_ServiceError(t *testing.T) {
	mock := &mockService{
		getUserFamiliesFn: func(ctx context.Context, userID string) ([]FamilyWithChildren, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("GET", "/families", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] != "database connection failed" {
		t.Errorf("Expected error message, got %s", response["error"])
	}
}

// ============================================================================
// Create Family Tests
// ============================================================================

func TestCreateFamily_Success(t *testing.T) {
	now := time.Now()
	mock := &mockService{
		createFamilyFn: func(ctx context.Context, userID string, req *CreateFamilyRequest) (*Family, error) {
			if userID != "test-user" {
				t.Errorf("Expected userID test-user, got %s", userID)
			}
			if req.Name != "New Family" {
				t.Errorf("Expected name New Family, got %s", req.Name)
			}
			return &Family{
				ID:        "new-family-id",
				Name:      req.Name,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"name": "New Family"}`
	req := httptest.NewRequest("POST", "/families", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response Family
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.ID != "new-family-id" {
		t.Errorf("Expected ID new-family-id, got %s", response.ID)
	}

	if response.Name != "New Family" {
		t.Errorf("Expected name New Family, got %s", response.Name)
	}
}

func TestCreateFamily_ValidationError_MissingName(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{}`
	req := httptest.NewRequest("POST", "/families", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateFamily_ValidationError_InvalidJSON(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{invalid json}`
	req := httptest.NewRequest("POST", "/families", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateFamily_ServiceError(t *testing.T) {
	mock := &mockService{
		createFamilyFn: func(ctx context.Context, userID string, req *CreateFamilyRequest) (*Family, error) {
			return nil, errors.New("failed to create family")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"name": "New Family"}`
	req := httptest.NewRequest("POST", "/families", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// Get Family Tests
// ============================================================================

func TestGetFamily_Success(t *testing.T) {
	now := time.Now()
	mock := &mockService{
		getFamilyFn: func(ctx context.Context, familyID string) (*Family, error) {
			if familyID != "family-123" {
				t.Errorf("Expected familyID family-123, got %s", familyID)
			}
			return &Family{
				ID:        familyID,
				Name:      "Test Family",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("GET", "/families/family-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Family
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.ID != "family-123" {
		t.Errorf("Expected ID family-123, got %s", response.ID)
	}
}

func TestGetFamily_ServiceError(t *testing.T) {
	mock := &mockService{
		getFamilyFn: func(ctx context.Context, familyID string) (*Family, error) {
			return nil, errors.New("family not found")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("GET", "/families/family-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// Update Family Tests
// ============================================================================

func TestUpdateFamily_Success(t *testing.T) {
	now := time.Now()
	mock := &mockService{
		updateFamilyFn: func(ctx context.Context, familyID string, req *CreateFamilyRequest) (*Family, error) {
			if familyID != "family-123" {
				t.Errorf("Expected familyID family-123, got %s", familyID)
			}
			if req.Name != "Updated Family Name" {
				t.Errorf("Expected name Updated Family Name, got %s", req.Name)
			}
			return &Family{
				ID:        familyID,
				Name:      req.Name,
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"name": "Updated Family Name"}`
	req := httptest.NewRequest("PUT", "/families/family-123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Family
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Name != "Updated Family Name" {
		t.Errorf("Expected name Updated Family Name, got %s", response.Name)
	}
}

func TestUpdateFamily_ValidationError_MissingName(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{}`
	req := httptest.NewRequest("PUT", "/families/family-123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateFamily_ValidationError_InvalidJSON(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{invalid json}`
	req := httptest.NewRequest("PUT", "/families/family-123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateFamily_ServiceError(t *testing.T) {
	mock := &mockService{
		updateFamilyFn: func(ctx context.Context, familyID string, req *CreateFamilyRequest) (*Family, error) {
			return nil, errors.New("family not found")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"name": "Updated Family Name"}`
	req := httptest.NewRequest("PUT", "/families/family-123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// Delete Family Tests
// ============================================================================

func TestDeleteFamily_Success(t *testing.T) {
	mock := &mockService{
		deleteFamilyFn: func(ctx context.Context, familyID, userID string) error {
			if familyID != "family-123" {
				t.Errorf("Expected familyID family-123, got %s", familyID)
			}
			if userID != "test-user" {
				t.Errorf("Expected userID test-user, got %s", userID)
			}
			return nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("DELETE", "/families/family-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestDeleteFamily_Forbidden_NotAdmin(t *testing.T) {
	mock := &mockService{
		deleteFamilyFn: func(ctx context.Context, familyID, userID string) error {
			return errors.New("only admins can delete a family")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("DELETE", "/families/family-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] != "only admins can delete a family" {
		t.Errorf("Expected error message, got %s", response["error"])
	}
}

func TestDeleteFamily_ServiceError(t *testing.T) {
	mock := &mockService{
		deleteFamilyFn: func(ctx context.Context, familyID, userID string) error {
			return errors.New("database error")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("DELETE", "/families/family-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestDeleteFamily_UsesCorrectUserID(t *testing.T) {
	var capturedUserID string
	mock := &mockService{
		deleteFamilyFn: func(ctx context.Context, familyID, userID string) error {
			capturedUserID = userID
			return nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouterWithUserID(handler, "admin-user-123")

	req := httptest.NewRequest("DELETE", "/families/family-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "admin-user-123" {
		t.Errorf("Expected userID admin-user-123, got %s", capturedUserID)
	}
}

// ============================================================================
// Leave Family Tests
// ============================================================================

func TestLeaveFamily_Success(t *testing.T) {
	mock := &mockService{
		leaveFamilyFn: func(ctx context.Context, familyID, userID string) error {
			if familyID != "family-123" {
				t.Errorf("Expected familyID family-123, got %s", familyID)
			}
			if userID != "test-user" {
				t.Errorf("Expected userID test-user, got %s", userID)
			}
			return nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("POST", "/families/family-123/leave", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestLeaveFamily_OnlyAdmin(t *testing.T) {
	mock := &mockService{
		leaveFamilyFn: func(ctx context.Context, familyID, userID string) error {
			return errors.New("cannot leave: you are the only admin")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("POST", "/families/family-123/leave", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["error"] != "cannot leave: you are the only admin" {
		t.Errorf("Expected error message, got %s", response["error"])
	}
}

func TestLeaveFamily_ServiceError(t *testing.T) {
	mock := &mockService{
		leaveFamilyFn: func(ctx context.Context, familyID, userID string) error {
			return errors.New("database error")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("POST", "/families/family-123/leave", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestLeaveFamily_UsesCorrectUserID(t *testing.T) {
	var capturedUserID string
	mock := &mockService{
		leaveFamilyFn: func(ctx context.Context, familyID, userID string) error {
			capturedUserID = userID
			return nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouterWithUserID(handler, "leaving-user-456")

	req := httptest.NewRequest("POST", "/families/family-123/leave", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "leaving-user-456" {
		t.Errorf("Expected userID leaving-user-456, got %s", capturedUserID)
	}
}

// ============================================================================
// List Members Tests
// ============================================================================

func TestListMembers_Success(t *testing.T) {
	now := time.Now()
	mock := &mockService{
		getFamilyMembersFn: func(ctx context.Context, familyID string) ([]MemberWithUser, error) {
			if familyID != "family-123" {
				t.Errorf("Expected familyID family-123, got %s", familyID)
			}
			return []MemberWithUser{
				{
					ID:        "member-1",
					UserID:    "user-1",
					Name:      "John Doe",
					Email:     "john@example.com",
					Role:      "admin",
					CreatedAt: now,
				},
				{
					ID:        "member-2",
					UserID:    "user-2",
					Name:      "Jane Doe",
					Email:     "jane@example.com",
					Role:      "member",
					CreatedAt: now,
				},
			}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("GET", "/families/family-123/members", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []MemberWithUser
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 members, got %d", len(response))
	}

	if response[0].Role != "admin" {
		t.Errorf("Expected role admin, got %s", response[0].Role)
	}
}

func TestListMembers_ServiceError(t *testing.T) {
	mock := &mockService{
		getFamilyMembersFn: func(ctx context.Context, familyID string) ([]MemberWithUser, error) {
			return nil, errors.New("failed to get members")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("GET", "/families/family-123/members", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// Invite Member Tests
// ============================================================================

func TestInviteMember_Success(t *testing.T) {
	mock := &mockService{
		inviteMemberFn: func(ctx context.Context, familyID string, req *InviteRequest) error {
			if familyID != "family-123" {
				t.Errorf("Expected familyID family-123, got %s", familyID)
			}
			if req.Email != "invite@example.com" {
				t.Errorf("Expected email invite@example.com, got %s", req.Email)
			}
			return nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"email": "invite@example.com"}`
	req := httptest.NewRequest("POST", "/families/family-123/invite", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["message"] != "invite sent" {
		t.Errorf("Expected message 'invite sent', got %s", response["message"])
	}
}

func TestInviteMember_ValidationError_MissingEmail(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{}`
	req := httptest.NewRequest("POST", "/families/family-123/invite", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestInviteMember_ValidationError_InvalidEmail(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"email": "not-an-email"}`
	req := httptest.NewRequest("POST", "/families/family-123/invite", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestInviteMember_ValidationError_InvalidJSON(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{invalid}`
	req := httptest.NewRequest("POST", "/families/family-123/invite", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestInviteMember_ServiceError(t *testing.T) {
	mock := &mockService{
		inviteMemberFn: func(ctx context.Context, familyID string, req *InviteRequest) error {
			return errors.New("failed to send invite")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"email": "invite@example.com"}`
	req := httptest.NewRequest("POST", "/families/family-123/invite", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// Join Family Tests
// ============================================================================

func TestJoinFamily_Success(t *testing.T) {
	now := time.Now()
	mock := &mockService{
		joinFamilyFn: func(ctx context.Context, familyID, userID string) (*Family, error) {
			if familyID != "family-123" {
				t.Errorf("Expected familyID family-123, got %s", familyID)
			}
			if userID != "test-user" {
				t.Errorf("Expected userID test-user, got %s", userID)
			}
			return &Family{
				ID:        familyID,
				Name:      "Joined Family",
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("POST", "/families/family-123/join", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response Family
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Name != "Joined Family" {
		t.Errorf("Expected name Joined Family, got %s", response.Name)
	}
}

func TestJoinFamily_NotFound(t *testing.T) {
	mock := &mockService{
		joinFamilyFn: func(ctx context.Context, familyID, userID string) (*Family, error) {
			return nil, errors.New("family not found")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("POST", "/families/nonexistent/join", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestJoinFamily_ServiceError(t *testing.T) {
	mock := &mockService{
		joinFamilyFn: func(ctx context.Context, familyID, userID string) (*Family, error) {
			return nil, errors.New("already a member")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("POST", "/families/family-123/join", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// Remove Member Tests
// ============================================================================

func TestRemoveMember_Success(t *testing.T) {
	mock := &mockService{
		removeMemberFn: func(ctx context.Context, familyID, userID string) error {
			if familyID != "family-123" {
				t.Errorf("Expected familyID family-123, got %s", familyID)
			}
			if userID != "user-456" {
				t.Errorf("Expected userID user-456, got %s", userID)
			}
			return nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("DELETE", "/families/family-123/members/user-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestRemoveMember_ServiceError(t *testing.T) {
	mock := &mockService{
		removeMemberFn: func(ctx context.Context, familyID, userID string) error {
			return errors.New("cannot remove last admin")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("DELETE", "/families/family-123/members/user-456", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// List Children Tests
// ============================================================================

func TestListChildren_Success(t *testing.T) {
	now := time.Now()
	dob := time.Date(2020, 5, 15, 0, 0, 0, 0, time.UTC)
	mock := &mockService{
		getChildrenFn: func(ctx context.Context, familyID string) ([]Child, error) {
			if familyID != "family-123" {
				t.Errorf("Expected familyID family-123, got %s", familyID)
			}
			return []Child{
				{
					ID:          "child-1",
					FamilyID:    familyID,
					Name:        "Alice",
					DateOfBirth: dob,
					Gender:      "female",
					CreatedAt:   now,
					UpdatedAt:   now,
				},
				{
					ID:          "child-2",
					FamilyID:    familyID,
					Name:        "Bob",
					DateOfBirth: dob,
					Gender:      "male",
					CreatedAt:   now,
					UpdatedAt:   now,
				},
			}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("GET", "/families/family-123/children", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []Child
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response) != 2 {
		t.Errorf("Expected 2 children, got %d", len(response))
	}

	if response[0].Name != "Alice" {
		t.Errorf("Expected name Alice, got %s", response[0].Name)
	}
}

func TestListChildren_Empty(t *testing.T) {
	mock := &mockService{
		getChildrenFn: func(ctx context.Context, familyID string) ([]Child, error) {
			return []Child{}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("GET", "/families/family-123/children", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response []Child
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if len(response) != 0 {
		t.Errorf("Expected 0 children, got %d", len(response))
	}
}

func TestListChildren_ServiceError(t *testing.T) {
	mock := &mockService{
		getChildrenFn: func(ctx context.Context, familyID string) ([]Child, error) {
			return nil, errors.New("failed to get children")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("GET", "/families/family-123/children", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// Add Child Tests
// ============================================================================

func TestAddChild_Success(t *testing.T) {
	now := time.Now()
	dob := time.Date(2020, 5, 15, 0, 0, 0, 0, time.UTC)
	mock := &mockService{
		addChildFn: func(ctx context.Context, familyID string, req *AddChildRequest) (*Child, error) {
			if familyID != "family-123" {
				t.Errorf("Expected familyID family-123, got %s", familyID)
			}
			if req.Name != "Charlie" {
				t.Errorf("Expected name Charlie, got %s", req.Name)
			}
			return &Child{
				ID:          "new-child-id",
				FamilyID:    familyID,
				Name:        req.Name,
				DateOfBirth: req.DateOfBirth,
				Gender:      req.Gender,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"name": "Charlie", "date_of_birth": "2020-05-15T00:00:00Z", "gender": "male"}`
	req := httptest.NewRequest("POST", "/families/family-123/children", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response Child
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.ID != "new-child-id" {
		t.Errorf("Expected ID new-child-id, got %s", response.ID)
	}

	if response.Name != "Charlie" {
		t.Errorf("Expected name Charlie, got %s", response.Name)
	}

	if response.DateOfBirth != dob {
		t.Errorf("Expected DOB %v, got %v", dob, response.DateOfBirth)
	}
}

func TestAddChild_ValidationError_MissingName(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"date_of_birth": "2020-05-15T00:00:00Z"}`
	req := httptest.NewRequest("POST", "/families/family-123/children", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAddChild_ValidationError_MissingDateOfBirth(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"name": "Charlie"}`
	req := httptest.NewRequest("POST", "/families/family-123/children", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAddChild_ValidationError_InvalidJSON(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{invalid json`
	req := httptest.NewRequest("POST", "/families/family-123/children", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestAddChild_ServiceError(t *testing.T) {
	mock := &mockService{
		addChildFn: func(ctx context.Context, familyID string, req *AddChildRequest) (*Child, error) {
			return nil, errors.New("failed to add child")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"name": "Charlie", "date_of_birth": "2020-05-15T00:00:00Z"}`
	req := httptest.NewRequest("POST", "/families/family-123/children", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// Update Child Tests
// ============================================================================

func TestUpdateChild_Success(t *testing.T) {
	now := time.Now()
	dob := time.Date(2020, 5, 15, 0, 0, 0, 0, time.UTC)
	mock := &mockService{
		updateChildFn: func(ctx context.Context, childID string, req *AddChildRequest) (*Child, error) {
			if childID != "child-123" {
				t.Errorf("Expected childID child-123, got %s", childID)
			}
			if req.Name != "Charlie Updated" {
				t.Errorf("Expected name Charlie Updated, got %s", req.Name)
			}
			return &Child{
				ID:          childID,
				FamilyID:    "family-123",
				Name:        req.Name,
				DateOfBirth: req.DateOfBirth,
				Gender:      req.Gender,
				CreatedAt:   now,
				UpdatedAt:   now,
			}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"name": "Charlie Updated", "date_of_birth": "2020-05-15T00:00:00Z", "gender": "male"}`
	req := httptest.NewRequest("PUT", "/families/family-123/children/child-123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var response Child
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Name != "Charlie Updated" {
		t.Errorf("Expected name Charlie Updated, got %s", response.Name)
	}

	if response.DateOfBirth != dob {
		t.Errorf("Expected DOB %v, got %v", dob, response.DateOfBirth)
	}
}

func TestUpdateChild_ValidationError_MissingName(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"date_of_birth": "2020-05-15T00:00:00Z"}`
	req := httptest.NewRequest("PUT", "/families/family-123/children/child-123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateChild_ValidationError_InvalidJSON(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{not valid}`
	req := httptest.NewRequest("PUT", "/families/family-123/children/child-123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestUpdateChild_ServiceError(t *testing.T) {
	mock := &mockService{
		updateChildFn: func(ctx context.Context, childID string, req *AddChildRequest) (*Child, error) {
			return nil, errors.New("child not found")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	body := `{"name": "Charlie Updated", "date_of_birth": "2020-05-15T00:00:00Z"}`
	req := httptest.NewRequest("PUT", "/families/family-123/children/child-123", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// Delete Child Tests
// ============================================================================

func TestDeleteChild_Success(t *testing.T) {
	mock := &mockService{
		deleteChildFn: func(ctx context.Context, childID string) error {
			if childID != "child-123" {
				t.Errorf("Expected childID child-123, got %s", childID)
			}
			return nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("DELETE", "/families/family-123/children/child-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestDeleteChild_ServiceError(t *testing.T) {
	mock := &mockService{
		deleteChildFn: func(ctx context.Context, childID string) error {
			return errors.New("failed to delete child")
		},
	}

	handler := NewHandler(mock)
	router := setupRouter(handler)

	req := httptest.NewRequest("DELETE", "/families/family-123/children/child-123", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

// ============================================================================
// Handler Construction Tests
// ============================================================================

func TestNewHandler(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)

	if handler == nil {
		t.Fatal("Expected handler to be non-nil")
	}

	if handler.service == nil {
		t.Error("Expected service to be set")
	}
}

func TestRegisterRoutes(t *testing.T) {
	mock := &mockService{}
	handler := NewHandler(mock)

	router := gin.New()
	families := router.Group("/families")
	handler.RegisterRoutes(families)

	// Verify routes are registered by making requests
	routes := router.Routes()

	expectedRoutes := map[string]string{
		"GET/families":                                "listFamilies",
		"POST/families":                               "createFamily",
		"GET/families/:familyId":                      "getFamily",
		"PUT/families/:familyId":                      "updateFamily",
		"DELETE/families/:familyId":                   "deleteFamily",
		"POST/families/:familyId/leave":               "leaveFamily",
		"GET/families/:familyId/members":              "listMembers",
		"POST/families/:familyId/invite":              "inviteMember",
		"POST/families/:familyId/join":                "joinFamily",
		"DELETE/families/:familyId/members/:userId":   "removeMember",
		"GET/families/:familyId/children":             "listChildren",
		"POST/families/:familyId/children":            "addChild",
		"PUT/families/:familyId/children/:childId":    "updateChild",
		"DELETE/families/:familyId/children/:childId": "deleteChild",
	}

	registeredRoutes := make(map[string]bool)
	for _, route := range routes {
		key := route.Method + route.Path
		registeredRoutes[key] = true
	}

	for key := range expectedRoutes {
		if !registeredRoutes[key] {
			t.Errorf("Expected route %s to be registered", key)
		}
	}
}

// ============================================================================
// Context User ID Tests
// ============================================================================

func TestListFamilies_UsesCorrectUserID(t *testing.T) {
	var capturedUserID string
	mock := &mockService{
		getUserFamiliesFn: func(ctx context.Context, userID string) ([]FamilyWithChildren, error) {
			capturedUserID = userID
			return []FamilyWithChildren{}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouterWithUserID(handler, "specific-user-123")

	req := httptest.NewRequest("GET", "/families", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "specific-user-123" {
		t.Errorf("Expected userID specific-user-123, got %s", capturedUserID)
	}
}

func TestCreateFamily_UsesCorrectUserID(t *testing.T) {
	var capturedUserID string
	mock := &mockService{
		createFamilyFn: func(ctx context.Context, userID string, req *CreateFamilyRequest) (*Family, error) {
			capturedUserID = userID
			return &Family{ID: "fam-1", Name: req.Name}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouterWithUserID(handler, "another-user-456")

	body := `{"name": "Test Family"}`
	req := httptest.NewRequest("POST", "/families", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "another-user-456" {
		t.Errorf("Expected userID another-user-456, got %s", capturedUserID)
	}
}

func TestJoinFamily_UsesCorrectUserID(t *testing.T) {
	var capturedUserID string
	mock := &mockService{
		joinFamilyFn: func(ctx context.Context, familyID, userID string) (*Family, error) {
			capturedUserID = userID
			return &Family{ID: familyID, Name: "Family"}, nil
		},
	}

	handler := NewHandler(mock)
	router := setupRouterWithUserID(handler, "joining-user-789")

	req := httptest.NewRequest("POST", "/families/family-123/join", http.NoBody)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if capturedUserID != "joining-user-789" {
		t.Errorf("Expected userID joining-user-789, got %s", capturedUserID)
	}
}
