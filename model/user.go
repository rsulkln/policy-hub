package model

type User struct {
	ID       string `bson:"_id,omitempty"`
	Name     string `bson:"username"`
	Password string `bson:"password"`
	Role     string `bson:"role"`
}
