package validation

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

var Validator *validator.Validate

const (
	ValidPassword = "valid_password"
)

func init() {
	Validator = validator.New()

	Validator.RegisterValidation(ValidPassword, func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		validPasswordRegex := regexp.MustCompile(`^[A-Za-z\d!@#$%^&*()_+{}\[\]:;<>,.?~\\-\\'\\"]{8,}$`)
		return validPasswordRegex.MatchString(password)
	})
}
