package validation

import (
	"github.com/go-playground/validator"
)

func Validate(item interface{}) map[string]string {
	var fieldErrors map[string]string = map[string]string{}
	val := validator.New()
	err := val.Struct(item)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			fieldErrors[err.Field()] = err.Tag()
		}
	}
	return fieldErrors
}
