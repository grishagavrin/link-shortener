package main

// nodemon --watch ../../ --exec go run main.go --signal SIGTERM
// go test ./... -v

import (
	"fmt"
	"log"
	"net/http"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/handlers"
)

func main() {

	fmt.Printf("Server startder on %s:%s", config.HOST, config.PORT)
	err := http.ListenAndServe(config.HOST+":"+config.PORT, MyHandler())
	if err != nil {
		log.Fatal("Could not start server: ", err)
	}
}

func MyHandler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/", handlers.CommonHandler)
	return r
}
