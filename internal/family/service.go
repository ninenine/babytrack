package family

import (
	"context"
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
	// TODO: implement
	return nil, nil
}

func (s *service) GetFamily(ctx context.Context, familyID string) (*Family, error) {
	return s.repo.GetFamilyByID(ctx, familyID)
}

func (s *service) GetUserFamilies(ctx context.Context, userID string) ([]Family, error) {
	return s.repo.GetUserFamilies(ctx, userID)
}

func (s *service) InviteMember(ctx context.Context, familyID string, req *InviteRequest) error {
	// TODO: implement - send invite email
	return nil
}

func (s *service) RemoveMember(ctx context.Context, familyID, userID string) error {
	return s.repo.RemoveFamilyMember(ctx, familyID, userID)
}

func (s *service) AddChild(ctx context.Context, familyID string, req *AddChildRequest) (*Child, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) GetChildren(ctx context.Context, familyID string) ([]Child, error) {
	return s.repo.GetChildren(ctx, familyID)
}

func (s *service) UpdateChild(ctx context.Context, childID string, req *AddChildRequest) (*Child, error) {
	// TODO: implement
	return nil, nil
}

func (s *service) DeleteChild(ctx context.Context, childID string) error {
	return s.repo.DeleteChild(ctx, childID)
}
