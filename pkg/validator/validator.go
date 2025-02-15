package validator

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()

	v.RegisterValidation("username", validateUsername)
	v.RegisterValidation("password", validatePassword)

	return &Validator{
		validate: v,
	}
}

func (v *Validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		var errs []string
		for _, err := range err.(validator.ValidationErrors) {
			errs = append(errs, formatError(err))
		}
		return fmt.Errorf(strings.Join(errs, "; "))
	}
	return nil
}

func formatError(err validator.FieldError) string {
	field := err.Field()
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, err.Param())
	case "username":
		return fmt.Sprintf("%s contains invalid characters", field)
	case "password":
		return fmt.Sprintf("%s is not strong enough", field)
	default:
		return fmt.Sprintf("%s is invalid", field)
	}
}

func validateUsername(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	return len(username) >= 3 && len(username) <= 50
}

func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	return len(password) >= 6
}
