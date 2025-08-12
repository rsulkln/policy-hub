package model

import "context"

type User struct {
	ID       string `bson:"_id,omitempty"`
	Name     string `bson:"username"`
	Password string `bson:"password"`
	Role     string `bson:"role"`
}
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByUsername(ctx context.Context, username string) (*User, error)
}
