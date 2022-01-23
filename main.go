package main

import (
	"log"
	"net/http"
	"site/database"
	"site/routes"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	server := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
	}

	connection, err := database.NewDatabaseConnection()
	if err != nil {
		log.Fatalln(err)
	}
	if err := database.RunMigrations(connection); err != nil {
		log.Fatalf("Error running migrations %s \n", err)
	}

	routes.NewRouteRegister(router)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalln(err)
	}
}
