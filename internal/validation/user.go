package validation

import (
	"flarecloud/internal/models"

	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)


type UserValidationError struct {
	Errors validator.ValidationErrors
}

func ValidateUserSignUp(u models.UserSignUp) error {
	err := Validator.Struct(u)
	if err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		
		if ok {
			userValidationError := UserValidationError{Errors: validationErrors}
			return userValidationError.Error(validationErrors)
		}

		return err
	}
	return nil
}

// func ValidateUserLogin(u models.UserLogin) error {
// 	err := Validator.Struct(u)
// 	if err != nil {
// 		validationErrors := err.(validator.ValidationErrors)
// 		return UserValidationError(validationErrors)
// 	}
// 	return nil
// }

// func ValidateUser(u models.User) error {
// 	err := Validator.Struct(u)
// 	if err != nil {
// 		validationErrors := err.(validator.ValidationErrors)
// 		return UserValidationError(validationErrors)
// 	}
// 	return nil
// }

func(e UserValidationError) Error(errors validator.ValidationErrors) error {
	errorMessages := make([]string, 0)
	for _, err := range errors {
		var errorMessage string
		separator := ",\n"

		if err.Field() == models.USER_FIELD_USERNAME {
			errorMessage = fmt.Sprintf("`%s` is not valid the value is %s%s", strings.ToLower(err.Field()), err.Tag(), separator)
		}

		if err.Field() == models.USER_FIELD_PASSWORD {
			errorMessage = fmt.Sprintf("`%s` is not valid please verify the password format it must have at least 8 characters, one uppercase, one lowercase, one number and one special character%s", strings.ToLower(err.Field()), separator)
		}

		if err.Field() == models.USER_FIELD_EMAIL {
			errorMessage = fmt.Sprintf("`%s` is not valid please verify the email format%s", strings.ToLower(err.Field()), separator)
		}

		errorMessages = append(errorMessages, errorMessage)
	}
	return fmt.Errorf(strings.Join(errorMessages, ""))
}