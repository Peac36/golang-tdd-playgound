package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"site/database"
	"site/routes"
	"syscall"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	server := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8080",
	}

	connection, err := database.NewDatabaseConnection()
	if err != nil {
		log.Fatalln(err)
	}

	defer database.CloseDatabaseConnection(connection)

	if err := database.RunMigrations(connection); err != nil {
		log.Fatalf("Error running migrations %s \n", err)
	}

	go func() {
		routes.NewRouteRegister(router)

		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalln(err)
		}
	}()

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT)
	<-signalCh
	//set a limit of 30 seconds before completely shutdown the server
	ctx, shutdown := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdown()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}
	log.Println("Graceful shutdown complete.")

}
