package fbr

import (
	"prom/app/db"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()


type ErrorResponse struct {
    FailedField string
    Tag         string
    Value       string
}

func ValidateStruct[T db.User](s T) []*ErrorResponse {
	var errors []*ErrorResponse
	err := validate.Struct(s)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.FailedField = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}
