package family

import (
	"context"
	"database/sql"
)

type Repository interface {
	// Family
	GetFamilyByID(ctx context.Context, id string) (*Family, error)
	CreateFamily(ctx context.Context, family *Family) error
	UpdateFamily(ctx context.Context, family *Family) error

	// Members
	GetFamilyMembers(ctx context.Context, familyID string) ([]FamilyMember, error)
	AddFamilyMember(ctx context.Context, member *FamilyMember) error
	RemoveFamilyMember(ctx context.Context, familyID, userID string) error
	GetUserFamilies(ctx context.Context, userID string) ([]Family, error)

	// Children
	GetChildren(ctx context.Context, familyID string) ([]Child, error)
	GetChildByID(ctx context.Context, id string) (*Child, error)
	CreateChild(ctx context.Context, child *Child) error
	UpdateChild(ctx context.Context, child *Child) error
	DeleteChild(ctx context.Context, id string) error
}

type repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetFamilyByID(ctx context.Context, id string) (*Family, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) CreateFamily(ctx context.Context, family *Family) error {
	// TODO: implement
	return nil
}

func (r *repository) UpdateFamily(ctx context.Context, family *Family) error {
	// TODO: implement
	return nil
}

func (r *repository) GetFamilyMembers(ctx context.Context, familyID string) ([]FamilyMember, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) AddFamilyMember(ctx context.Context, member *FamilyMember) error {
	// TODO: implement
	return nil
}

func (r *repository) RemoveFamilyMember(ctx context.Context, familyID, userID string) error {
	// TODO: implement
	return nil
}

func (r *repository) GetUserFamilies(ctx context.Context, userID string) ([]Family, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) GetChildren(ctx context.Context, familyID string) ([]Child, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) GetChildByID(ctx context.Context, id string) (*Child, error) {
	// TODO: implement
	return nil, nil
}

func (r *repository) CreateChild(ctx context.Context, child *Child) error {
	// TODO: implement
	return nil
}

func (r *repository) UpdateChild(ctx context.Context, child *Child) error {
	// TODO: implement
	return nil
}

func (r *repository) DeleteChild(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}
