package family

import (
	"context"
	"database/sql"
	"errors"
)

type Repository interface {
	// Family
	GetFamilyByID(ctx context.Context, id string) (*Family, error)
	CreateFamily(ctx context.Context, family *Family) error
	UpdateFamily(ctx context.Context, family *Family) error

	// Members
	GetFamilyMembers(ctx context.Context, familyID string) ([]FamilyMember, error)
	GetFamilyMembersWithUsers(ctx context.Context, familyID string) ([]MemberWithUser, error)
	AddFamilyMember(ctx context.Context, member *FamilyMember) error
	RemoveFamilyMember(ctx context.Context, familyID, userID string) error
	GetUserFamilies(ctx context.Context, userID string) ([]Family, error)
	IsMember(ctx context.Context, familyID, userID string) (bool, error)

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

// Family methods

func (r *repository) GetFamilyByID(ctx context.Context, id string) (*Family, error) {
	query := `SELECT id, name, created_at, updated_at FROM families WHERE id = $1`

	var family Family
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&family.ID,
		&family.Name,
		&family.CreatedAt,
		&family.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &family, nil
}

func (r *repository) CreateFamily(ctx context.Context, family *Family) error {
	query := `INSERT INTO families (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query,
		family.ID,
		family.Name,
		family.CreatedAt,
		family.UpdatedAt,
	)

	return err
}

func (r *repository) UpdateFamily(ctx context.Context, family *Family) error {
	query := `UPDATE families SET name = $2, updated_at = $3 WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query,
		family.ID,
		family.Name,
		family.UpdatedAt,
	)

	return err
}

// Member methods

func (r *repository) GetFamilyMembers(ctx context.Context, familyID string) ([]FamilyMember, error) {
	query := `SELECT id, family_id, user_id, role, created_at FROM family_members WHERE family_id = $1`

	rows, err := r.db.QueryContext(ctx, query, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []FamilyMember
	for rows.Next() {
		var m FamilyMember
		if err := rows.Scan(&m.ID, &m.FamilyID, &m.UserID, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return members, rows.Err()
}

func (r *repository) GetFamilyMembersWithUsers(ctx context.Context, familyID string) ([]MemberWithUser, error) {
	query := `
		SELECT fm.id, fm.user_id, u.name, u.email, u.avatar_url, fm.role, fm.created_at
		FROM family_members fm
		INNER JOIN users u ON fm.user_id = u.id
		WHERE fm.family_id = $1
		ORDER BY fm.created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []MemberWithUser
	for rows.Next() {
		var m MemberWithUser
		var avatarURL sql.NullString

		if err := rows.Scan(&m.ID, &m.UserID, &m.Name, &m.Email, &avatarURL, &m.Role, &m.CreatedAt); err != nil {
			return nil, err
		}

		if avatarURL.Valid {
			m.AvatarURL = avatarURL.String
		}

		members = append(members, m)
	}

	return members, rows.Err()
}

func (r *repository) AddFamilyMember(ctx context.Context, member *FamilyMember) error {
	query := `INSERT INTO family_members (id, family_id, user_id, role, created_at) VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(ctx, query,
		member.ID,
		member.FamilyID,
		member.UserID,
		member.Role,
		member.CreatedAt,
	)

	return err
}

func (r *repository) RemoveFamilyMember(ctx context.Context, familyID, userID string) error {
	query := `DELETE FROM family_members WHERE family_id = $1 AND user_id = $2`

	_, err := r.db.ExecContext(ctx, query, familyID, userID)
	return err
}

func (r *repository) IsMember(ctx context.Context, familyID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM family_members WHERE family_id = $1 AND user_id = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, familyID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (r *repository) GetUserFamilies(ctx context.Context, userID string) ([]Family, error) {
	query := `
		SELECT f.id, f.name, f.created_at, f.updated_at
		FROM families f
		INNER JOIN family_members fm ON f.id = fm.family_id
		WHERE fm.user_id = $1
		ORDER BY f.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var families []Family
	for rows.Next() {
		var f Family
		if err := rows.Scan(&f.ID, &f.Name, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		families = append(families, f)
	}

	return families, rows.Err()
}

// Children methods

func (r *repository) GetChildren(ctx context.Context, familyID string) ([]Child, error) {
	query := `
		SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at
		FROM children
		WHERE family_id = $1
		ORDER BY date_of_birth DESC
	`

	rows, err := r.db.QueryContext(ctx, query, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var children []Child
	for rows.Next() {
		var c Child
		var gender, avatarURL sql.NullString

		if err := rows.Scan(
			&c.ID, &c.FamilyID, &c.Name, &c.DateOfBirth,
			&gender, &avatarURL, &c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if gender.Valid {
			c.Gender = gender.String
		}
		if avatarURL.Valid {
			c.AvatarURL = avatarURL.String
		}

		children = append(children, c)
	}

	return children, rows.Err()
}

func (r *repository) GetChildByID(ctx context.Context, id string) (*Child, error) {
	query := `
		SELECT id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at
		FROM children
		WHERE id = $1
	`

	var c Child
	var gender, avatarURL sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.FamilyID, &c.Name, &c.DateOfBirth,
		&gender, &avatarURL, &c.CreatedAt, &c.UpdatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if gender.Valid {
		c.Gender = gender.String
	}
	if avatarURL.Valid {
		c.AvatarURL = avatarURL.String
	}

	return &c, nil
}

func (r *repository) CreateChild(ctx context.Context, child *Child) error {
	query := `
		INSERT INTO children (id, family_id, name, date_of_birth, gender, avatar_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	var gender, avatarURL *string
	if child.Gender != "" {
		gender = &child.Gender
	}
	if child.AvatarURL != "" {
		avatarURL = &child.AvatarURL
	}

	_, err := r.db.ExecContext(ctx, query,
		child.ID,
		child.FamilyID,
		child.Name,
		child.DateOfBirth,
		gender,
		avatarURL,
		child.CreatedAt,
		child.UpdatedAt,
	)

	return err
}

func (r *repository) UpdateChild(ctx context.Context, child *Child) error {
	query := `
		UPDATE children
		SET name = $2, date_of_birth = $3, gender = $4, avatar_url = $5, updated_at = $6
		WHERE id = $1
	`

	var gender, avatarURL *string
	if child.Gender != "" {
		gender = &child.Gender
	}
	if child.AvatarURL != "" {
		avatarURL = &child.AvatarURL
	}

	_, err := r.db.ExecContext(ctx, query,
		child.ID,
		child.Name,
		child.DateOfBirth,
		gender,
		avatarURL,
		child.UpdatedAt,
	)

	return err
}

func (r *repository) DeleteChild(ctx context.Context, id string) error {
	query := `DELETE FROM children WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
