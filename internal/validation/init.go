package validation

import (
	"log"
	"regexp"

	"github.com/go-playground/validator/v10"
)

var Validator *validator.Validate

const (
	ValidPassword = "valid_password"
)

func init() {
	Validator = validator.New()

	err := Validator.RegisterValidation(ValidPassword, func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		validPasswordRegex := regexp.MustCompile(`^[A-Za-z\d!@#$%^&*()_+{}\[\]:;<>,.?~\\-\\'\\"]{8,}$`)
		return validPasswordRegex.MatchString(password)
	})

	if err != nil {
		log.Fatal(err)
	}
}
