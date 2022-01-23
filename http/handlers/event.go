package handlers

import (
	"encoding/json"
	"net/http"
	"regexp"
	"site/database/models"
	"site/http/middlewares"
	"site/http/responses"
	"site/security"
	"site/validation"
	"strconv"

	"gorm.io/gorm"
)

func EventCreate(connection *gorm.DB, tokenService security.TokenSecurity) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responses.NewJsonResponse(rw, http.StatusMethodNotAllowed, nil)
			return
		}

		event := models.Event{}
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, nil)
			return
		}

		errors := validation.Validate(event)
		if len(errors) != 0 {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, errors)
			return
		}

		authHeader := r.Header.Get(middlewares.AuthorizationHeader)
		if authHeader == "" {
			responses.NewJsonResponse(rw, http.StatusUnauthorized, nil)
			return
		}

		userId, err := tokenService.GetIdentifier(authHeader)
		if err != nil {
			responses.NewJsonResponse(rw, http.StatusUnauthorized, nil)
			return
		}

		user := models.User{}
		connection.Find(&user, userId)
		event.UserID = user.ID

		result := connection.Save(&event)
		if result.Error != nil {
			responses.NewJsonResponse(rw, http.StatusInternalServerError, nil)
			return
		}
		responses.NewJsonResponse(rw, http.StatusOK, nil)
	})
}

func GetEvent(connection *gorm.DB) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responses.NewJsonResponse(rw, http.StatusMethodNotAllowed, nil)
			return
		}

		regex := regexp.MustCompile(`\/event\/(\d+)\/?$`)
		regexResult := regex.FindAllStringSubmatch(r.URL.Path, 1)
		eventIdString := regexResult[0][1]
		eventId, err := strconv.Atoi(eventIdString)

		if err != nil || eventId == 0 {
			responses.NewJsonResponse(rw, http.StatusNotFound, nil)
			return
		}

		event := models.Event{}
		result := connection.Find(&event, eventId)
		if result.Error != nil {
			responses.NewJsonResponse(rw, http.StatusInternalServerError, nil)
			return
		}

		if event.ID == 0 {
			responses.NewJsonResponse(rw, http.StatusNotFound, nil)
			return
		}

		responses.NewJsonResponse(rw, http.StatusOK, event)

	})
}

func GetEvents(connection *gorm.DB, t security.TokenSecurity) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			responses.NewJsonResponse(rw, http.StatusMethodNotAllowed, nil)
			return
		}

		token := r.Header.Get(middlewares.AuthorizationHeader)
		identifier, err := t.GetIdentifier(token)
		if err != nil {
			responses.NewJsonResponse(rw, http.StatusUnauthorized, nil)
			return
		}

		user := models.User{}
		result := connection.Preload("Events").Find(&user, identifier)
		if result.Error != nil {
			responses.NewJsonResponse(rw, http.StatusInternalServerError, nil)
			return
		}

		if user.ID == 0 {
			responses.NewJsonResponse(rw, http.StatusNotFound, nil)
			return
		}

		responses.NewJsonResponse(rw, http.StatusOK, user.Events)
	})
}
