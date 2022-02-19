package handlers

import (
	"log"
	"net/http"
	"os"
	"site/database/models"
	"site/http/middlewares"
	"site/http/responses"
	"site/security"
	"site/uploader"
	"strconv"

	"gorm.io/gorm"
)

func CreateMedia(connection *gorm.DB, security security.TokenSecurity, uploaderService uploader.Uploader) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			responses.NewJsonResponse(rw, http.StatusMethodNotAllowed, nil)
			return
		}

		eventId, err := ParseEventId(r)
		if err != nil {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, nil)
			return
		}

		event := models.Event{}
		result := connection.Find(&event, eventId)
		if result.Error != nil || event.ID != 0 {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, map[string]string{
				"error": "The event can be found!",
			})
			return
		}

		identifier := r.Header.Get(middlewares.AuthorizationHeader)
		userId, err := security.GetIdentifier(identifier)
		if err != nil {
			responses.NewJsonResponse(rw, http.StatusUnauthorized, nil)
			return
		}

		user := models.User{}
		result = connection.Find(&user, userId)
		if result.Error != nil {
			responses.NewJsonResponse(rw, http.StatusUnauthorized, nil)
			return
		}

		if event.UserID != user.ID {
			responses.NewJsonResponse(rw, http.StatusForbidden, nil)
			return
		}

		r.ParseMultipartForm(10 << 20)
		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, nil)
			return
		}

		dir, err := os.Getwd()
		if err != nil {
			responses.NewJsonResponse(rw, http.StatusInternalServerError, nil)
			return
		}

		path, err := uploaderService.Upload("test.jpg", dir+"/files/", file)
		if err != nil {
			responses.NewJsonResponse(rw, http.StatusUnprocessableEntity, map[string]string{
				"error": "File can not be uploaded",
			})
			log.Println(err)
			return
		}

		media := models.Media{
			Name:     fileHeader.Filename,
			Size:     int(fileHeader.Size),
			Order:    0,
			Provider: "local",
			Path:     path,
			EventId:  uint(eventId),
		}

		result = connection.Save(&media)
		if result.Error != nil {
			responses.NewJsonResponse(rw, http.StatusInternalServerError, nil)
			return
		}

		responses.NewJsonResponse(rw, http.StatusOK, map[string]string{
			"media_id": strconv.Itoa(int(media.ID)),
			"event_id": strconv.Itoa(int(eventId)),
		})
	})
}
