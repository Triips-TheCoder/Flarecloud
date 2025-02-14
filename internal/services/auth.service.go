package services

import (
	"context"
	"errors"
	"flarecloud/internal/models"
	"flarecloud/internal/validation"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

type AuthService struct {
	userCollection *mongo.Collection
}


func NewAuthService(userCollection *mongo.Collection) *AuthService {
	return &AuthService{
		userCollection: userCollection,
	}
}

func (a *AuthService) SignUp(ctx context.Context, user models.UserSignUp) error {
	if err := validation.ValidateUserSignUp(user); err != nil {
		return err
	}

	
	count, err := a.userCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		return err
	}
	if count == 1 {
		return errors.New("email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}




	newUser := models.User{
		Username: user.Username,
		Email: user.Email,
		Password: string(hashedPassword),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}


	_, err = a.userCollection.InsertOne(ctx, newUser)

	return err
}

func (a *AuthService) Login(ctx context.Context, userInput models.UserLogin) (string, error) {

	if err := validation.ValidateUserLogin(userInput); err != nil {
		return "", err
	}


	var userFromDatabase models.User
	err := a.userCollection.FindOne(ctx, bson.M{"email": userInput.Email}).Decode(&userFromDatabase)
	if err != nil {
		return "", errors.New("email not found")
	}

	
	if err = bcrypt.CompareHashAndPassword([]byte(userFromDatabase.Password), []byte(userInput.Password)); err != nil {
		return "", errors.New("invalid credentials")
	}


	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userFromDatabase.ID.Hex(),
		"exp":     time.Now().Add(72 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}


func (a *AuthService) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token, nil
}