package domain

import "context"

type Role string

const (
	RolePublic  Role = "Public" // public access, token unnecessary
	RoleAdmin   Role = "Admin"
	RoleManager Role = "Manager"
	RoleUser    Role = "User"
)

type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"-"`
	Role     Role   `json:"role"`
}

type UserRepository interface {
	Save(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetAll(ctx context.Context) ([]User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type UserUseCase interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetAll(ctx context.Context) ([]User, error)
}
