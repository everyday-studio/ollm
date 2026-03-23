package domain

import (
	"context"
	"time"
)

type Role string

const (
	RolePublic  Role = "Public" // public access, token unnecessary
	RoleAdmin   Role = "Admin"
	RoleManager Role = "Manager"
	RoleUser    Role = "User"
)

type User struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Tag       string    `json:"tag"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	GoogleID  string    `json:"-"`
	Role      Role       `json:"role"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type UpdateNicknameRequest struct {
	Name string `json:"name"`
}

type UserRepository interface {
	Save(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetPaginated(ctx context.Context, page, limit int) ([]User, error)
	CountAll(ctx context.Context) (int, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	UpdateNickname(ctx context.Context, id string, name string) error
	// UpsertGoogleUser finds a user by google_id or creates a new one.
	// Returns the existing or newly created user.
	UpsertGoogleUser(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id string) error
}

type UserUseCase interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetPaginated(ctx context.Context, page, limit int) (*PaginatedData[User], error)
	CountAll(ctx context.Context) (int, error)
	UpdateNickname(ctx context.Context, id string, name string) (*User, error)
	Delete(ctx context.Context, id string) error
}
