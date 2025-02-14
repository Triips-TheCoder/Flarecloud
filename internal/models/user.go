package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	USER_FIELD_PASSWORD = "Password"
	USER_FIELD_USERNAME = "Username"
	USER_FIELD_EMAIL    = "Email"
)




type UserSignUp struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,valid_password"`
	CreatedAt time.Time `json:"created_at"`
}


type UserLogin struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username  string             `bson:"username" json:"username"`
	Email     string             `bson:"email" json:"email"`
	Password  string             `bson:"password" json:"password"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}