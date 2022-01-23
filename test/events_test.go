package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"site/database"
	"site/database/models"
	"site/http/handlers"
	"site/http/middlewares"
	"site/security"
	"testing"

	"gorm.io/gorm"
)

func TestGetEvents(t *testing.T) {
	sec := security.NewTokenService()
	t.Run("it_returns_methon_not_allowed", func(t *testing.T) {
		connection, err := database.NewTestDatabaseConnection()
		if err != nil {
			t.Error("Can not get db connection")
		}
		database.RunMigrations(connection)

		r, err := http.NewRequest(http.MethodPost, "/event", nil)
		if err != nil {
			t.Errorf("Can not create a request: %s", err)
		}
		rw := httptest.NewRecorder()

		handlers.GetEvents(connection, sec).ServeHTTP(rw, r)
		if rw.Code != http.StatusMethodNotAllowed {
			t.Errorf("Unexpected status code. Received: %d, Expected: %d", rw.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("it_returns_unauthorized_if_token_is_missing", func(t *testing.T) {
		connection, err := database.NewTestDatabaseConnection()
		if err != nil {
			t.Error("Can not get db connection")
		}
		database.RunMigrations(connection)

		r, err := http.NewRequest(http.MethodGet, "/event", nil)
		if err != nil {
			t.Errorf("Can not create a request: %s", err)
		}
		rw := httptest.NewRecorder()

		handlers.GetEvents(connection, sec).ServeHTTP(rw, r)
		if rw.Code != http.StatusUnauthorized {
			t.Errorf("Unexpected status code. Received: %d, Expected: %d", rw.Code, http.StatusUnauthorized)
		}
	})

	t.Run("it_returns_not_found_if_user_is_not_found", func(t *testing.T) {
		connection, err := database.NewTestDatabaseConnection()
		if err != nil {
			t.Error("Can not get db connection")
		}
		database.RunMigrations(connection)

		user := models.User{
			Email:    "example@example.com",
			Password: "123456789",
			Model: gorm.Model{
				ID: 123456,
			},
		}

		token, err := sec.CreateToken(&user)
		if err != nil {
			t.Errorf("Can not create a user token %s", err)
		}

		r, err := http.NewRequest(http.MethodGet, "/event", nil)
		if err != nil {
			t.Errorf("Can not create a request: %s", err)
		}
		r.Header.Add(middlewares.AuthorizationHeader, token)
		rw := httptest.NewRecorder()

		handlers.GetEvents(connection, sec).ServeHTTP(rw, r)
		if rw.Code != http.StatusNotFound {
			t.Errorf("Unexpected status code. Received: %d, Expected: %d", rw.Code, http.StatusNotFound)
		}
	})

	t.Run("it_returns_user_events", func(t *testing.T) {
		connection, err := database.NewTestDatabaseConnection()
		if err != nil {
			t.Error("Can not get db connection")
		}
		database.RunMigrations(connection)

		user := models.User{
			Email:    "example1@example.com",
			Password: "123456789",
		}

		result := connection.Save(&user)
		if result.Error != nil {
			t.Errorf("Can not insert a user; %s", result.Error)
		}
		var events []models.Event = []models.Event{
			{Name: "Test Event1", UserID: user.ID},
			{Name: "Test Event2", UserID: user.ID},
			{Name: "Test Event3", UserID: user.ID},
		}

		result = connection.Create(&events)
		if result.Error != nil {
			t.Errorf("Can not store user events: %s", result.Error)
		}

		token, err := sec.CreateToken(&user)
		if err != nil {
			t.Errorf("Can not create a user token %s", err)
		}

		r, err := http.NewRequest(http.MethodGet, "/event", nil)
		if err != nil {
			t.Errorf("Can not create a request: %s", err)
		}

		r.Header.Add(middlewares.AuthorizationHeader, token)
		rw := httptest.NewRecorder()

		handlers.GetEvents(connection, sec).ServeHTTP(rw, r)

		if rw.Code != http.StatusOK {
			t.Errorf("Unexpected status code. Received: %d, Expected: %d", rw.Code, http.StatusOK)
		}

		var responseEvents []models.Event
		err = json.NewDecoder(rw.Body).Decode(&responseEvents)
		if err != nil {
			t.Errorf("Can not parse respones body %s", err)
		}

		if len(responseEvents) != 3 {
			t.Errorf("Incorrect event count %d", len(responseEvents))
		}

		if responseEvents[0].ID != events[0].ID {
			t.Error("Incorrect event returned")
		}
	})
}
