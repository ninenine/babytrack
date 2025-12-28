package family

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockRepository is a test double for Repository
type mockRepository struct {
	families        map[string]*Family
	members         map[string][]FamilyMember
	children        map[string]*Child
	userFamilies    map[string][]Family
	createFamilyErr error
	addMemberErr    error
	createChildErr  error
	updateChildErr  error
	deleteChildErr  error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		families:     make(map[string]*Family),
		members:      make(map[string][]FamilyMember),
		children:     make(map[string]*Child),
		userFamilies: make(map[string][]Family),
	}
}

func (m *mockRepository) GetFamilyByID(ctx context.Context, id string) (*Family, error) {
	f, ok := m.families[id]
	if !ok {
		return nil, nil
	}
	return f, nil
}

func (m *mockRepository) CreateFamily(ctx context.Context, family *Family) error {
	if m.createFamilyErr != nil {
		return m.createFamilyErr
	}
	m.families[family.ID] = family
	return nil
}

func (m *mockRepository) UpdateFamily(ctx context.Context, family *Family) error {
	m.families[family.ID] = family
	return nil
}

func (m *mockRepository) GetFamilyMembers(ctx context.Context, familyID string) ([]FamilyMember, error) {
	return m.members[familyID], nil
}

func (m *mockRepository) GetFamilyMembersWithUsers(ctx context.Context, familyID string) ([]MemberWithUser, error) {
	members := m.members[familyID]
	result := make([]MemberWithUser, len(members))
	for i, member := range members {
		result[i] = MemberWithUser{
			ID:     member.ID,
			UserID: member.UserID,
			Role:   member.Role,
		}
	}
	return result, nil
}

func (m *mockRepository) AddFamilyMember(ctx context.Context, member *FamilyMember) error {
	if m.addMemberErr != nil {
		return m.addMemberErr
	}
	m.members[member.FamilyID] = append(m.members[member.FamilyID], *member)
	// Also add to user families
	if f, ok := m.families[member.FamilyID]; ok {
		m.userFamilies[member.UserID] = append(m.userFamilies[member.UserID], *f)
	}
	return nil
}

func (m *mockRepository) RemoveFamilyMember(ctx context.Context, familyID, userID string) error {
	members := m.members[familyID]
	for i, member := range members {
		if member.UserID == userID {
			m.members[familyID] = append(members[:i], members[i+1:]...)
			break
		}
	}
	return nil
}

func (m *mockRepository) GetUserFamilies(ctx context.Context, userID string) ([]Family, error) {
	return m.userFamilies[userID], nil
}

func (m *mockRepository) IsMember(ctx context.Context, familyID, userID string) (bool, error) {
	for _, member := range m.members[familyID] {
		if member.UserID == userID {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockRepository) GetChildren(ctx context.Context, familyID string) ([]Child, error) {
	var result []Child
	for _, child := range m.children {
		if child.FamilyID == familyID {
			result = append(result, *child)
		}
	}
	return result, nil
}

func (m *mockRepository) GetChildByID(ctx context.Context, id string) (*Child, error) {
	c, ok := m.children[id]
	if !ok {
		return nil, nil
	}
	return c, nil
}

func (m *mockRepository) CreateChild(ctx context.Context, child *Child) error {
	if m.createChildErr != nil {
		return m.createChildErr
	}
	m.children[child.ID] = child
	return nil
}

func (m *mockRepository) UpdateChild(ctx context.Context, child *Child) error {
	if m.updateChildErr != nil {
		return m.updateChildErr
	}
	m.children[child.ID] = child
	return nil
}

func (m *mockRepository) DeleteChild(ctx context.Context, id string) error {
	if m.deleteChildErr != nil {
		return m.deleteChildErr
	}
	delete(m.children, id)
	return nil
}

func TestService_CreateFamily(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &CreateFamilyRequest{
		Name: "Smith Family",
	}

	family, err := svc.CreateFamily(context.Background(), "user-123", req)
	if err != nil {
		t.Fatalf("CreateFamily() error = %v", err)
	}

	if family.ID == "" {
		t.Error("CreateFamily() should generate an ID")
	}

	if family.Name != req.Name {
		t.Errorf("CreateFamily() Name = %v, want %v", family.Name, req.Name)
	}

	// Check that creator was added as admin
	members := repo.members[family.ID]
	if len(members) != 1 {
		t.Fatalf("CreateFamily() should add creator as member, got %d members", len(members))
	}

	if members[0].Role != "admin" {
		t.Errorf("CreateFamily() creator role = %v, want admin", members[0].Role)
	}

	if members[0].UserID != "user-123" {
		t.Errorf("CreateFamily() creator UserID = %v, want user-123", members[0].UserID)
	}
}

func TestService_CreateFamily_RepoError(t *testing.T) {
	repo := newMockRepository()
	repo.createFamilyErr = errors.New("database error")
	svc := NewService(repo)

	req := &CreateFamilyRequest{
		Name: "Smith Family",
	}

	_, err := svc.CreateFamily(context.Background(), "user-123", req)
	if err == nil {
		t.Error("CreateFamily() should return error when repo fails")
	}
}

func TestService_CreateFamily_AddMemberError(t *testing.T) {
	repo := newMockRepository()
	repo.addMemberErr = errors.New("database error")
	svc := NewService(repo)

	req := &CreateFamilyRequest{
		Name: "Smith Family",
	}

	_, err := svc.CreateFamily(context.Background(), "user-123", req)
	if err == nil {
		t.Error("CreateFamily() should return error when adding member fails")
	}
}

func TestService_GetFamily(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a family first
	family := &Family{
		ID:        "family-123",
		Name:      "Test Family",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.families[family.ID] = family

	// Get it back
	retrieved, err := svc.GetFamily(context.Background(), family.ID)
	if err != nil {
		t.Fatalf("GetFamily() error = %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetFamily() returned nil for existing family")
	}

	if retrieved.ID != family.ID {
		t.Errorf("GetFamily() ID = %v, want %v", retrieved.ID, family.ID)
	}
}

func TestService_GetFamily_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	family, err := svc.GetFamily(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetFamily() error = %v", err)
	}

	if family != nil {
		t.Error("GetFamily() should return nil for non-existent family")
	}
}

func TestService_GetUserFamilies(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create family and add user
	family := &Family{
		ID:        "family-123",
		Name:      "Test Family",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.families[family.ID] = family
	repo.userFamilies["user-123"] = []Family{*family}

	// Get user's families
	families, err := svc.GetUserFamilies(context.Background(), "user-123")
	if err != nil {
		t.Fatalf("GetUserFamilies() error = %v", err)
	}

	if len(families) != 1 {
		t.Fatalf("GetUserFamilies() returned %d families, want 1", len(families))
	}

	if families[0].ID != family.ID {
		t.Errorf("GetUserFamilies() family ID = %v, want %v", families[0].ID, family.ID)
	}
}

func TestService_GetUserFamilies_Empty(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	families, err := svc.GetUserFamilies(context.Background(), "user-no-families")
	if err != nil {
		t.Fatalf("GetUserFamilies() error = %v", err)
	}

	if families == nil {
		t.Error("GetUserFamilies() should return empty slice, not nil")
	}

	if len(families) != 0 {
		t.Errorf("GetUserFamilies() returned %d families, want 0", len(families))
	}
}

func TestService_JoinFamily(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a family
	family := &Family{
		ID:        "family-123",
		Name:      "Test Family",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.families[family.ID] = family

	// Join the family
	joined, err := svc.JoinFamily(context.Background(), family.ID, "user-456")
	if err != nil {
		t.Fatalf("JoinFamily() error = %v", err)
	}

	if joined.ID != family.ID {
		t.Errorf("JoinFamily() returned family ID = %v, want %v", joined.ID, family.ID)
	}

	// Check that user was added as member
	members := repo.members[family.ID]
	found := false
	for _, m := range members {
		if m.UserID == "user-456" && m.Role == "member" {
			found = true
			break
		}
	}
	if !found {
		t.Error("JoinFamily() should add user as member")
	}
}

func TestService_JoinFamily_AlreadyMember(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a family with existing member
	family := &Family{
		ID:        "family-123",
		Name:      "Test Family",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.families[family.ID] = family
	repo.members[family.ID] = []FamilyMember{
		{ID: "member-1", FamilyID: family.ID, UserID: "user-123", Role: "admin"},
	}

	// Try to join again
	joined, err := svc.JoinFamily(context.Background(), family.ID, "user-123")
	if err != nil {
		t.Fatalf("JoinFamily() error = %v", err)
	}

	// Should just return the family without error
	if joined.ID != family.ID {
		t.Errorf("JoinFamily() should return family for existing member")
	}

	// Should not duplicate member
	if len(repo.members[family.ID]) != 1 {
		t.Errorf("JoinFamily() should not duplicate member, got %d members", len(repo.members[family.ID]))
	}
}

func TestService_JoinFamily_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	_, err := svc.JoinFamily(context.Background(), "non-existent", "user-123")
	if err == nil {
		t.Error("JoinFamily() should return error for non-existent family")
	}
}

func TestService_AddChild(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	dob := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	req := &AddChildRequest{
		Name:        "Baby Smith",
		DateOfBirth: dob,
		Gender:      "female",
	}

	child, err := svc.AddChild(context.Background(), "family-123", req)
	if err != nil {
		t.Fatalf("AddChild() error = %v", err)
	}

	if child.ID == "" {
		t.Error("AddChild() should generate an ID")
	}

	if child.FamilyID != "family-123" {
		t.Errorf("AddChild() FamilyID = %v, want family-123", child.FamilyID)
	}

	if child.Name != req.Name {
		t.Errorf("AddChild() Name = %v, want %v", child.Name, req.Name)
	}

	if !child.DateOfBirth.Equal(dob) {
		t.Errorf("AddChild() DateOfBirth = %v, want %v", child.DateOfBirth, dob)
	}

	if child.Gender != req.Gender {
		t.Errorf("AddChild() Gender = %v, want %v", child.Gender, req.Gender)
	}
}

func TestService_AddChild_RepoError(t *testing.T) {
	repo := newMockRepository()
	repo.createChildErr = errors.New("database error")
	svc := NewService(repo)

	req := &AddChildRequest{
		Name:        "Baby Smith",
		DateOfBirth: time.Now(),
	}

	_, err := svc.AddChild(context.Background(), "family-123", req)
	if err == nil {
		t.Error("AddChild() should return error when repo fails")
	}
}

func TestService_GetChild(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a child
	child := &Child{
		ID:          "child-123",
		FamilyID:    "family-123",
		Name:        "Baby Smith",
		DateOfBirth: time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo.children[child.ID] = child

	// Get it back
	retrieved, err := svc.GetChild(context.Background(), child.ID)
	if err != nil {
		t.Fatalf("GetChild() error = %v", err)
	}

	if retrieved == nil {
		t.Fatal("GetChild() returned nil for existing child")
	}

	if retrieved.ID != child.ID {
		t.Errorf("GetChild() ID = %v, want %v", retrieved.ID, child.ID)
	}
}

func TestService_GetChild_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	child, err := svc.GetChild(context.Background(), "non-existent")
	if err != nil {
		t.Fatalf("GetChild() error = %v", err)
	}

	if child != nil {
		t.Error("GetChild() should return nil for non-existent child")
	}
}

func TestService_UpdateChild(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a child
	child := &Child{
		ID:          "child-123",
		FamilyID:    "family-123",
		Name:        "Old Name",
		DateOfBirth: time.Now().AddDate(-1, 0, 0),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	repo.children[child.ID] = child

	// Update it
	newDOB := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC)
	req := &AddChildRequest{
		Name:        "New Name",
		DateOfBirth: newDOB,
		Gender:      "male",
	}

	updated, err := svc.UpdateChild(context.Background(), child.ID, req)
	if err != nil {
		t.Fatalf("UpdateChild() error = %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("UpdateChild() Name = %v, want New Name", updated.Name)
	}

	if !updated.DateOfBirth.Equal(newDOB) {
		t.Errorf("UpdateChild() DateOfBirth = %v, want %v", updated.DateOfBirth, newDOB)
	}
}

func TestService_UpdateChild_NotFound(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	req := &AddChildRequest{
		Name:        "New Name",
		DateOfBirth: time.Now(),
	}

	_, err := svc.UpdateChild(context.Background(), "non-existent", req)
	if err == nil {
		t.Error("UpdateChild() should return error for non-existent child")
	}
}

func TestService_DeleteChild(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create a child
	child := &Child{
		ID:       "child-123",
		FamilyID: "family-123",
		Name:     "Baby Smith",
	}
	repo.children[child.ID] = child

	// Delete it
	err := svc.DeleteChild(context.Background(), child.ID)
	if err != nil {
		t.Fatalf("DeleteChild() error = %v", err)
	}

	// Verify it's gone
	if _, ok := repo.children[child.ID]; ok {
		t.Error("DeleteChild() should remove the child")
	}
}

func TestService_GetChildren(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Create children
	child1 := &Child{ID: "child-1", FamilyID: "family-123", Name: "Child 1"}
	child2 := &Child{ID: "child-2", FamilyID: "family-123", Name: "Child 2"}
	child3 := &Child{ID: "child-3", FamilyID: "family-456", Name: "Child 3"} // Different family
	repo.children[child1.ID] = child1
	repo.children[child2.ID] = child2
	repo.children[child3.ID] = child3

	// Get children for family-123
	children, err := svc.GetChildren(context.Background(), "family-123")
	if err != nil {
		t.Fatalf("GetChildren() error = %v", err)
	}

	if len(children) != 2 {
		t.Errorf("GetChildren() returned %d children, want 2", len(children))
	}
}

func TestService_GetChildren_Empty(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	children, err := svc.GetChildren(context.Background(), "family-no-children")
	if err != nil {
		t.Fatalf("GetChildren() error = %v", err)
	}

	if children == nil {
		t.Error("GetChildren() should return empty slice, not nil")
	}

	if len(children) != 0 {
		t.Errorf("GetChildren() returned %d children, want 0", len(children))
	}
}

func TestService_RemoveMember(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Add a member
	repo.members["family-123"] = []FamilyMember{
		{ID: "member-1", FamilyID: "family-123", UserID: "user-123", Role: "member"},
		{ID: "member-2", FamilyID: "family-123", UserID: "user-456", Role: "admin"},
	}

	// Remove one
	err := svc.RemoveMember(context.Background(), "family-123", "user-123")
	if err != nil {
		t.Fatalf("RemoveMember() error = %v", err)
	}

	// Verify only one remains
	if len(repo.members["family-123"]) != 1 {
		t.Errorf("RemoveMember() should leave 1 member, got %d", len(repo.members["family-123"]))
	}

	if repo.members["family-123"][0].UserID != "user-456" {
		t.Error("RemoveMember() removed wrong member")
	}
}

func TestService_GetFamilyMembers(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	// Add members
	repo.members["family-123"] = []FamilyMember{
		{ID: "member-1", FamilyID: "family-123", UserID: "user-123", Role: "admin"},
		{ID: "member-2", FamilyID: "family-123", UserID: "user-456", Role: "member"},
	}

	members, err := svc.GetFamilyMembers(context.Background(), "family-123")
	if err != nil {
		t.Fatalf("GetFamilyMembers() error = %v", err)
	}

	if len(members) != 2 {
		t.Errorf("GetFamilyMembers() returned %d members, want 2", len(members))
	}
}

func TestService_GetFamilyMembers_Empty(t *testing.T) {
	repo := newMockRepository()
	svc := NewService(repo)

	members, err := svc.GetFamilyMembers(context.Background(), "family-no-members")
	if err != nil {
		t.Fatalf("GetFamilyMembers() error = %v", err)
	}

	if members == nil {
		t.Error("GetFamilyMembers() should return empty slice, not nil")
	}
}
