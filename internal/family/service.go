package family

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type Service interface {
	// Family
	CreateFamily(ctx context.Context, userID string, req *CreateFamilyRequest) (*Family, error)
	GetFamily(ctx context.Context, familyID string) (*Family, error)
	GetUserFamilies(ctx context.Context, userID string) ([]FamilyWithChildren, error)

	// Members
	GetFamilyMembers(ctx context.Context, familyID string) ([]MemberWithUser, error)
	InviteMember(ctx context.Context, familyID string, req *InviteRequest) error
	JoinFamily(ctx context.Context, familyID, userID string) (*Family, error)
	RemoveMember(ctx context.Context, familyID, userID string) error

	// Children
	AddChild(ctx context.Context, familyID string, req *AddChildRequest) (*Child, error)
	GetChildren(ctx context.Context, familyID string) ([]Child, error)
	GetChild(ctx context.Context, childID string) (*Child, error)
	UpdateChild(ctx context.Context, childID string, req *AddChildRequest) (*Child, error)
	DeleteChild(ctx context.Context, childID string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateFamily(ctx context.Context, userID string, req *CreateFamilyRequest) (*Family, error) {
	now := time.Now()

	family := &Family{
		ID:        generateID(),
		Name:      req.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.CreateFamily(ctx, family); err != nil {
		return nil, fmt.Errorf("failed to create family: %w", err)
	}

	// Add the creator as admin
	member := &FamilyMember{
		ID:        generateID(),
		FamilyID:  family.ID,
		UserID:    userID,
		Role:      "admin",
		CreatedAt: now,
	}

	if err := s.repo.AddFamilyMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to add family member: %w", err)
	}

	return family, nil
}

func (s *service) GetFamily(ctx context.Context, familyID string) (*Family, error) {
	return s.repo.GetFamilyByID(ctx, familyID)
}

func (s *service) GetUserFamilies(ctx context.Context, userID string) ([]FamilyWithChildren, error) {
	families, err := s.repo.GetUserFamilies(ctx, userID)
	if err != nil {
		return nil, err
	}
	if families == nil {
		return []FamilyWithChildren{}, nil
	}

	// Fetch children for each family
	result := make([]FamilyWithChildren, len(families))
	for i, f := range families {
		children, err := s.repo.GetChildren(ctx, f.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get children for family %s: %w", f.ID, err)
		}
		if children == nil {
			children = []Child{}
		}
		result[i] = FamilyWithChildren{
			ID:        f.ID,
			Name:      f.Name,
			Children:  children,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
		}
	}
	return result, nil
}

func (s *service) GetFamilyMembers(ctx context.Context, familyID string) ([]MemberWithUser, error) {
	members, err := s.repo.GetFamilyMembersWithUsers(ctx, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family members: %w", err)
	}
	if members == nil {
		return []MemberWithUser{}, nil
	}
	return members, nil
}

func (s *service) InviteMember(ctx context.Context, familyID string, req *InviteRequest) error {
	// TODO: implement email invite
	return fmt.Errorf("not implemented")
}

func (s *service) JoinFamily(ctx context.Context, familyID, userID string) (*Family, error) {
	// Check if family exists
	family, err := s.repo.GetFamilyByID(ctx, familyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get family: %w", err)
	}
	if family == nil {
		return nil, fmt.Errorf("family not found")
	}

	// Check if user is already a member
	isMember, err := s.repo.IsMember(ctx, familyID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check membership: %w", err)
	}
	if isMember {
		return family, nil // Already a member, just return the family
	}

	// Add user as member
	member := &FamilyMember{
		ID:        generateID(),
		FamilyID:  familyID,
		UserID:    userID,
		Role:      "member",
		CreatedAt: time.Now(),
	}

	if err := s.repo.AddFamilyMember(ctx, member); err != nil {
		return nil, fmt.Errorf("failed to join family: %w", err)
	}

	return family, nil
}

func (s *service) RemoveMember(ctx context.Context, familyID, userID string) error {
	return s.repo.RemoveFamilyMember(ctx, familyID, userID)
}

func (s *service) AddChild(ctx context.Context, familyID string, req *AddChildRequest) (*Child, error) {
	now := time.Now()

	child := &Child{
		ID:          generateID(),
		FamilyID:    familyID,
		Name:        req.Name,
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateChild(ctx, child); err != nil {
		return nil, fmt.Errorf("failed to create child: %w", err)
	}

	return child, nil
}

func (s *service) GetChildren(ctx context.Context, familyID string) ([]Child, error) {
	children, err := s.repo.GetChildren(ctx, familyID)
	if err != nil {
		return nil, err
	}
	if children == nil {
		return []Child{}, nil
	}
	return children, nil
}

func (s *service) GetChild(ctx context.Context, childID string) (*Child, error) {
	return s.repo.GetChildByID(ctx, childID)
}

func (s *service) UpdateChild(ctx context.Context, childID string, req *AddChildRequest) (*Child, error) {
	child, err := s.repo.GetChildByID(ctx, childID)
	if err != nil {
		return nil, err
	}
	if child == nil {
		return nil, fmt.Errorf("child not found")
	}

	child.Name = req.Name
	child.DateOfBirth = req.DateOfBirth
	child.Gender = req.Gender
	child.UpdatedAt = time.Now()

	if err := s.repo.UpdateChild(ctx, child); err != nil {
		return nil, fmt.Errorf("failed to update child: %w", err)
	}

	return child, nil
}

func (s *service) DeleteChild(ctx context.Context, childID string) error {
	return s.repo.DeleteChild(ctx, childID)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck // crypto/rand.Read rarely fails
	return hex.EncodeToString(b)
}
