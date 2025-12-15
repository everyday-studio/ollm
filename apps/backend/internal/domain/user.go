package domain

import "context"

const (
	RolePublic  = "Public" // Token Unnecessary
	RoleAdmin   = "Admin"
	RoleManager = "Manager"
	RoleUser    = "User"
)

type User struct {
	ID       int64    `json:"id"`
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password string   `json:"-"`
	Roles    []string `json:"roles"`
}

type UserRepository interface {
	Save(user *User) (*User, error)
	GetByID(id int64) (*User, error)
	GetAll() ([]User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type UserUseCase interface {
	GetByID(id int64) (*User, error)
	GetAll() ([]User, error)
}
