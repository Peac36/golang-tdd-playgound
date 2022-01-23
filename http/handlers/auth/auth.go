package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"site/database/models"
	"site/http/responses"
	"site/security"
	"site/validation"

	"github.com/go-playground/validator"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type PossibleUser struct {
	Name                 string `validate:"required,min=6,email"`
	Password             string `validate:"required,min=6,eqcsfield=PasswordConfirmation"`
	PasswordConfirmation string `validate:"required,min=6"`
}

type UserLogin struct {
	Name     string `validate:"required"`
	Password string `validate:"required"`
}

func Register(connection *gorm.DB) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var user PossibleUser
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, struct{}{})
			return
		}

		validationErrors := validation.Validate(user)
		if len(validationErrors) != 0 {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, validationErrors)
			return
		}

		userCheck := models.User{}
		connection.Find(&userCheck, "email=?", user.Name)

		if userCheck.ID != 0 {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, map[string]string{
				"error": "The email is already taken!",
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalln("Failed to hash the password")
		}

		modelUser := models.User{
			Email:    user.Name,
			Password: string(hashedPassword),
		}

		result := connection.Save(&modelUser)
		if result.Error != nil {
			responses.NewJsonResponse(rw, http.StatusInternalServerError, nil)
			return
		}

		responses.NewJsonResponse(rw, http.StatusOK, modelUser)
	})
}

func Login(connection *gorm.DB, t security.TokenSecurity) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responses.NewJsonResponse(rw, http.StatusMethodNotAllowed, struct{}{})
			return
		}

		var user UserLogin

		json.NewDecoder(r.Body).Decode(&user)

		val := validator.New()
		err := val.Struct(user)
		if err != nil {
			fieldErrors := map[string]string{}
			for _, err := range err.(validator.ValidationErrors) {
				fieldErrors[err.Field()] = err.Tag()
			}

			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, fieldErrors)
			return
		}

		userCheck := models.User{}
		connection.Find(&userCheck, "email=?", user.Name)

		if userCheck.ID == 0 {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, map[string]string{
				"error": "Invalid User!",
			})
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(userCheck.Password), []byte(user.Password))

		if err != nil {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, map[string]string{
				"error": "Invalid User!",
			})
			return
		}

		tokenString, err := t.CreateToken(&userCheck)
		if err != nil {
			responses.NewJsonResponse(rw, http.StatusInternalServerError, nil)
			return
		}

		responses.NewJsonResponse(rw, http.StatusOK, map[string]string{
			"token": tokenString,
		})
	})
}
