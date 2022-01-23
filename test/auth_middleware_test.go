package test

import (
	"net/http"
	"net/http/httptest"
	"site/database"
	"site/database/models"
	"site/http/middlewares"
	"site/security"
	"testing"
)

func TestAuthMiddleware(t *testing.T) {
	middleware := middlewares.AuthMiddleware(security.NewTokenService())

	t.Run("it_returns_unauthorized_status_code_if_authorization_header_is_missing", func(t *testing.T) {
		r, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Errorf("Can not create a request %s", err)
		}
		rw := httptest.NewRecorder()

		// dummy handler
		d := http.NotFoundHandler()

		middleware(d).ServeHTTP(rw, r)

		if rw.Code != http.StatusUnauthorized {
			t.Errorf("Unexpected status code. Received %d; Expected %d", rw.Code, http.StatusUnauthorized)
		}
	})

	t.Run("it_returns_unauthorized_status_code_if_authorization_header_is_invalid", func(t *testing.T) {
		r, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Errorf("Can not create a request %s", err)
		}
		r.Header.Set(middlewares.AuthorizationHeader, "test")
		rw := httptest.NewRecorder()

		// dummy handler
		d := http.NotFoundHandler()

		middleware(d).ServeHTTP(rw, r)

		if rw.Code != http.StatusUnauthorized {
			t.Errorf("Unexpected status code. Received %d; Expected %d", rw.Code, http.StatusUnauthorized)
		}
	})

	t.Run("it_calls_the_next_middleware", func(t *testing.T) {
		connection, _ := database.NewTestDatabaseConnection()
		s := security.NewTokenService()

		database.RunMigrations(connection)

		user := models.User{
			Email:    "example@example.com",
			Password: "123456789",
		}

		result := connection.Save(&user)
		if result.Error != nil {
			t.Errorf("Can not store an user %s", result.Error)
		}
		token, err := s.CreateToken(&user)
		if err != nil {
			t.Errorf("Can not generate a token %s", err)
		}

		r, err := http.NewRequest(http.MethodGet, "/test", nil)
		if err != nil {
			t.Errorf("Can not create a request %s", err)
		}

		r.Header.Set(middlewares.AuthorizationHeader, token)
		rw := httptest.NewRecorder()

		// dummy handler
		d := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusOK)
		})

		middleware(d).ServeHTTP(rw, r)

		if rw.Code != http.StatusOK {
			t.Errorf("Unexpected status code. Received %d; Expected %d", rw.Code, http.StatusOK)
		}
	})

}
