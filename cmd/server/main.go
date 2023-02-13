package main

// nodemon --watch ../../ --exec go run main.go --signal SIGTERM

import (
	"fmt"
	"log"
	"net/http"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/handlers"
)

func main() {

	http.HandleFunc("/", handlers.CommonHandler)

	server := &http.Server{
		Addr: fmt.Sprintf("%s:%s", config.HOST, config.PORT),
	}

	fmt.Printf("Server startder on %s:%s", config.HOST, config.PORT)
	log.Fatal(server.ListenAndServe())

}
