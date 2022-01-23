package handlers

import (
	"net/http"
	"site/database/models"
	"site/http/middlewares"
	"site/http/responses"
	"site/security"

	"gorm.io/gorm"
)

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
