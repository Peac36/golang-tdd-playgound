package main

import (
	"log"
	"net/http"
	"site/database"
	"site/routes"
)

func main() {
	server := http.NewServeMux()

	connection, err := database.NewDatabaseConnection()
	if err != nil {
		log.Fatalln(err)
	}
	if err := database.RunMigrations(connection); err != nil {
		log.Fatalf("Error running migrations %s \n", err)
	}

	routes.NewRouteRegister(server)

	if err := http.ListenAndServe(":8080", server); err != nil {
		log.Fatalln(err)
	}
}
