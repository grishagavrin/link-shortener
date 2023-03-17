package main

// nodemon --watch ../../ --exec go run main.go --signal SIGTERM
// go test ./... -v

import (
	"fmt"
	"log"
	"net/http"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/routes"
)

func main() {
	config.SetFlag()
	cfg := config.ConfigENV{}
	serv, exists := cfg.GetEnvValue(config.ServerAddress)
	if !exists {
		log.Fatalf("env tag is not created, %s", config.ServerAddress)
	}

	fmt.Printf("Server started on %s", serv)
	err := http.ListenAndServe(serv, routes.ServiceRouter())
	if err != nil {
		log.Fatal("Could not start server: ", err)
	}
}
