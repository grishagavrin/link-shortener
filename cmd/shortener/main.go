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
	cfg := config.Instance()
	serv, err := cfg.GetCfgValue(config.ServerAddress)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("Server started on %s\n", serv)

	err = http.ListenAndServe(serv, routes.ServiceRouter())
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
