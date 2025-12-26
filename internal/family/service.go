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
	GetUserFamilies(ctx context.Context, userID string) ([]Family, error)

	// Members
	InviteMember(ctx context.Context, familyID string, req *InviteRequest) error
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

func (s *service) GetUserFamilies(ctx context.Context, userID string) ([]Family, error) {
	families, err := s.repo.GetUserFamilies(ctx, userID)
	if err != nil {
		return nil, err
	}
	if families == nil {
		return []Family{}, nil
	}
	return families, nil
}

func (s *service) InviteMember(ctx context.Context, familyID string, req *InviteRequest) error {
	// TODO: implement email invite
	return fmt.Errorf("not implemented")
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
	rand.Read(b)
	return hex.EncodeToString(b)
}
