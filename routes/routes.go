package routes

import (
	"site/database"
	"site/http/handlers"
	"site/http/handlers/auth"
	"site/http/middlewares"
	"site/security"

	"github.com/gorilla/mux"
)

func NewRouteRegister(server *mux.Router) {
	connection, _ := database.NewDatabaseConnection()
	tokenService := security.NewTokenService()

	authMiddleware := middlewares.AuthMiddleware(tokenService)

	server.Handle("/register", auth.Register(connection))
	server.Handle("/login", auth.Login(connection, tokenService))

	server.Handle("/event", authMiddleware(handlers.EventCreate(connection, tokenService)))
	server.Handle("/event/{event}", authMiddleware(handlers.GetEvent(connection)))
	server.Handle("/events", authMiddleware(handlers.GetEvents(connection, tokenService)))
}
