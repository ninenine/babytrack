package family

import "time"

type Family struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FamilyWithChildren struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Children  []Child   `json:"children"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FamilyMember struct {
	ID        string    `json:"id"`
	FamilyID  string    `json:"family_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"` // admin, member
	CreatedAt time.Time `json:"created_at"`
}

type Child struct {
	ID          string    `json:"id"`
	FamilyID    string    `json:"family_id"`
	Name        string    `json:"name"`
	DateOfBirth time.Time `json:"date_of_birth"`
	Gender      string    `json:"gender,omitempty"`
	AvatarURL   string    `json:"avatar_url,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateFamilyRequest struct {
	Name string `json:"name" binding:"required"`
}

type AddChildRequest struct {
	Name        string    `json:"name" binding:"required"`
	DateOfBirth time.Time `json:"date_of_birth" binding:"required"`
	Gender      string    `json:"gender,omitempty"`
}

type InviteRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// MemberWithUser includes user details for API responses
type MemberWithUser struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"joined_at"`
}
