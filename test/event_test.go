package test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"site/database"
	"site/database/models"
	"site/http/handlers"
	"strconv"
	"strings"
	"testing"
)

func TestCreateEvent(t *testing.T) {
	t.Run("it_allows_only_post_method", func(t *testing.T) {
		connection, err := database.NewTestDatabaseConnection()
		if err != nil {
			t.Error("Can not get db connection")
		}
		database.RunMigrations(connection)

		r, err := http.NewRequest(http.MethodGet, "event", nil)
		if err != nil {
			t.Error("Can not create a request")
		}
		rw := httptest.NewRecorder()

		handlers.EventCreate(connection).ServeHTTP(rw, r)

		if rw.Code != http.StatusMethodNotAllowed {
			t.Error("Incorrect method")
		}
	})

	t.Run("it_validates_input_data", func(t *testing.T) {
		connection, err := database.NewTestDatabaseConnection()
		if err != nil {
			t.Error("Can not get db connection")
		}
		database.RunMigrations(connection)

		event := models.Event{}
		jsonEvent, err := json.Marshal(event)
		if err != nil {
			t.Error("Can not encode request body")
		}

		r, err := http.NewRequest(http.MethodPost, "event", strings.NewReader(string(jsonEvent)))
		if err != nil {
			t.Error("Can not create a request")
		}
		rw := httptest.NewRecorder()

		handlers.EventCreate(connection).ServeHTTP(rw, r)

		if rw.Code != http.StatusUnprocessableEntity {
			t.Errorf("Unexpected status code. Expected: %d, Received %d", http.StatusUnprocessableEntity, rw.Code)
		}
	})

	t.Run("it_stores_the_event_data", func(t *testing.T) {
		connection, err := database.NewTestDatabaseConnection()
		if err != nil {
			t.Error("Can not get db connection")
		}
		database.RunMigrations(connection)

		event := models.Event{
			Name: "TestEvent123456",
		}
		jsonEvent, err := json.Marshal(event)
		if err != nil {
			t.Error("Can not encode request body")
		}

		r, err := http.NewRequest(http.MethodPost, "event", strings.NewReader(string(jsonEvent)))
		if err != nil {
			t.Error("Can not create a request")
		}
		rw := httptest.NewRecorder()

		storedEvent := models.Event{}
		result := connection.Find(&storedEvent, "Name=?", event.Name)

		if result.RowsAffected != 0 {
			t.Errorf("There is events with that name %s", event.Name)
		}

		handlers.EventCreate(connection).ServeHTTP(rw, r)

		storedEvent = models.Event{}
		result = connection.Find(&storedEvent, "Name=?", event.Name)

		if result.RowsAffected != 1 {
			t.Errorf("The event hasn't been stored\n")
		}

		if rw.Code != http.StatusOK {
			t.Errorf("There is a problem saving the event")
		}
	})

}

func TestGetEvent(t *testing.T) {
	t.Run("it_allows_only_get_method", func(t *testing.T) {
		connection, err := database.NewTestDatabaseConnection()
		if err != nil {
			t.Error("Can not get db connection")
		}
		database.RunMigrations(connection)

		r, err := http.NewRequest(http.MethodPost, "/event/1", nil)
		if err != nil {
			t.Errorf("Can not create an error %s", err)
		}
		rw := httptest.NewRecorder()

		handlers.GetEvent(connection).ServeHTTP(rw, r)

		if rw.Code != http.StatusMethodNotAllowed {
			t.Errorf("Unexpected status code. Received: %d, Expected: %d", rw.Code, http.StatusMethodNotAllowed)
		}

	})

	t.Run("it_returns_not_found_whenever_event can not be found", func(t *testing.T) {
		connection, err := database.NewTestDatabaseConnection()
		if err != nil {
			t.Error("Can not get db connection")
		}
		database.RunMigrations(connection)

		r, err := http.NewRequest(http.MethodGet, "/event/55", nil)
		if err != nil {
			t.Errorf("Can not create an error %s", err)
		}
		rw := httptest.NewRecorder()

		handlers.GetEvent(connection).ServeHTTP(rw, r)

		if rw.Code != http.StatusNotFound {
			t.Errorf("Unexpected status code. Received: %d, Expected: %d", rw.Code, http.StatusMethodNotAllowed)
		}
	})

	t.Run("it_returns_event", func(t *testing.T) {
		connection, err := database.NewTestDatabaseConnection()
		if err != nil {
			t.Error("Can not get db connection")
		}
		database.RunMigrations(connection)

		event := models.Event{Name: "TestEvent"}
		result := connection.Save(&event)
		if result.Error != nil {
			t.Errorf("Can not create an event %s", result.Error)
		}

		eventId := strconv.Itoa(int(event.ID))
		r, err := http.NewRequest("GET", "/event/"+eventId, nil)
		if err != nil {
			t.Error("Can not create a request")
		}
		rw := httptest.NewRecorder()

		handlers.GetEvent(connection).ServeHTTP(rw, r)

		if rw.Code != http.StatusOK {
			t.Errorf("Unexpected status code. Received: %d, Expected: %d", rw.Code, http.StatusOK)
		}

		responseEvent := models.Event{}
		json.NewDecoder(rw.Body).Decode(&responseEvent)

		if event.ID != responseEvent.ID {
			t.Errorf("Unexpected event. Received: %d, Expected: %d", responseEvent.ID, event.ID)
		}
	})
}
