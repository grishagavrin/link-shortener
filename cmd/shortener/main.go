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
	fmt.Printf("Server started on %s:%s", config.HOST, config.PORT)
	err := http.ListenAndServe(config.HOST+":"+config.PORT, routes.ServiceRouter())
	if err != nil {
		log.Fatal("Could not start server: ", err)
	}
}
