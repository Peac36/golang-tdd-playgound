package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"site/database"
	"site/database/models"
	"site/http/handlers/auth"
	"site/security"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestRegisterHandler(t *testing.T) {
	t.Run("user_can_not_be_created_without_data", func(t *testing.T) {
		body, _ := json.Marshal(auth.PossibleUser{})

		req, _ := http.NewRequest("POST", "/register", strings.NewReader(string(body)))
		rw := httptest.NewRecorder()

		connection, _ := database.NewTestDatabaseConnection()
		auth.Register(connection).ServeHTTP(rw, req)

		if rw.Code != http.StatusUnprocessableEntity {
			t.Errorf("Validation rules are skipped. Receive %d", rw.Code)
		}
		errors := map[string]string{}
		json.Unmarshal(rw.Body.Bytes(), &errors)

		if errors["Name"] == "" || errors["Password"] == "" || errors["PasswordConfirmation"] == "" {
			t.Error("Invalid validation error message type")
		}
	})

	t.Run("user_can_not_be_created_if_email_is_already_taken", func(t *testing.T) {
		connection, _ := database.NewTestDatabaseConnection()
		database.RunMigrations(connection)

		connection.Create(&models.User{
			Email:    "example@example.com",
			Password: "test",
		})

		body, _ := json.Marshal(auth.PossibleUser{
			Name:                 "example@example.com",
			Password:             "test123",
			PasswordConfirmation: "test123",
		})

		req, _ := http.NewRequest("POST", "/register", strings.NewReader(string(body)))
		rw := httptest.NewRecorder()

		auth.Register(connection).ServeHTTP(rw, req)

		if rw.Code != http.StatusUnprocessableEntity {
			t.Errorf("Invalid response code: %d, expected %d", rw.Code, http.StatusUnprocessableEntity)
		}
		errors := map[string]string{}
		json.Unmarshal(rw.Body.Bytes(), &errors)

		if errors["error"] != "The email is already taken!" {
			t.Error("Invalid validation error message type")
		}
	})

	t.Run("it_stores_an_user_into_the_database", func(t *testing.T) {
		connection, _ := database.NewTestDatabaseConnection()
		database.RunMigrations(connection)
		email := "example@example123.com"

		body, _ := json.Marshal(auth.PossibleUser{
			Name:                 email,
			Password:             "test123",
			PasswordConfirmation: "test123",
		})

		req, _ := http.NewRequest("POST", "/register", strings.NewReader(string(body)))
		rw := httptest.NewRecorder()

		user := models.User{}
		result := connection.Find(&user, "email=?", email)
		if result.RowsAffected != 0 {
			t.Errorf("There is already an user with email %s", email)
		}
		auth.Register(connection).ServeHTTP(rw, req)

		if rw.Code != http.StatusOK {
			t.Errorf("Invalid response code: %d, expected %d", rw.Code, http.StatusOK)
		}

		result = connection.Find(&user, "email=?", email)
		if result.RowsAffected != 1 {
			t.Errorf("Expects only one user with email %s", email)
		}
	})

}
func TestLoginHandler(t *testing.T) {
	t.Run("it_validates_the_incoming_data", func(t *testing.T) {
		connection, _ := database.NewTestDatabaseConnection()
		database.RunMigrations(connection)
		user := auth.UserLogin{}

		requestPayload, err := json.Marshal(user)
		if err != nil {
			t.Error("Can not encode json payload")
		}

		r, err := http.NewRequest("POST", "/login", strings.NewReader(string(requestPayload)))
		if err != nil {
			t.Error("Can not create a http request")
		}

		rw := httptest.NewRecorder()

		auth.Login(connection, security.TokenService{}).ServeHTTP(rw, r)

		if rw.Code != http.StatusUnprocessableEntity {
			t.Errorf("Unexpected response status code. Expected %d, received: %d", http.StatusUnprocessableEntity, rw.Code)
		}

		responeBody := map[string]string{}
		json.Unmarshal(rw.Body.Bytes(), &responeBody)

		if responeBody["Name"] == "" || responeBody["Password"] == "" {
			t.Errorf("Expected error messages are not present")
		}
	})

	t.Run("it_returns_error_if_there_is_no_user_with_these_credentials", func(t *testing.T) {
		connection, _ := database.NewTestDatabaseConnection()
		database.RunMigrations(connection)
		user := auth.UserLogin{
			Name:     "example123@example.com",
			Password: "1234567890",
		}

		requestPayload, err := json.Marshal(user)
		if err != nil {
			t.Error("Can not encode json payload")
		}

		r, err := http.NewRequest("POST", "/login", strings.NewReader(string(requestPayload)))
		if err != nil {
			t.Error("Can not create a http request")
		}

		rw := httptest.NewRecorder()

		auth.Login(connection, security.NewTokenService()).ServeHTTP(rw, r)

		if rw.Code != http.StatusUnprocessableEntity {
			t.Errorf("Unexpected response status code. Expected %d, received: %d", http.StatusUnprocessableEntity, rw.Code)
		}

		responeBody := map[string]string{}
		json.Unmarshal(rw.Body.Bytes(), &responeBody)

		if responeBody["error"] != "Invalid User!" {
			t.Errorf("Expected error messages are not present")
		}
	})

	t.Run("an_user_can_login_into_the_system", func(t *testing.T) {
		connection, _ := database.NewTestDatabaseConnection()
		database.RunMigrations(connection)
		user := auth.UserLogin{
			Name:     "example123@example.com",
			Password: "1234567890",
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			t.Errorf("Can not hash the password string")
		}
		connection.Save(&models.User{
			Email:    user.Name,
			Password: string(hashedPassword),
		})

		requestPayload, err := json.Marshal(user)
		if err != nil {
			t.Error("Can not encode json payload")
		}

		r, err := http.NewRequest("POST", "/login", strings.NewReader(string(requestPayload)))
		if err != nil {
			t.Error("Can not create a http request")
		}

		rw := httptest.NewRecorder()

		auth.Login(connection, security.NewTokenService()).ServeHTTP(rw, r)

		if rw.Code != http.StatusOK {
			t.Errorf("Unexpected response status code. Expected %d, received: %d", http.StatusOK, rw.Code)
		}

		result := map[string]string{}
		json.NewDecoder(rw.Body).Decode(&result)
		token, ok := result["token"]

		if !ok || token == "" {
			t.Error("The token is not present in the response")
		}
	})
}
