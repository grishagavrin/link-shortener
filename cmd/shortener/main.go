package main

//macOS -  nodemon --watch ../../ --exec go run main.go --signal SIGTERM
//windows -  nodemon --watch ../../ --exec go run main.go --signal SIGKILL
// go test ./... -v

// shortenertest -test.v -test.run=^TestIteration7$ -binary-path=cmd/shortener/shortener -server-port=8080 -file-storage-path=internal/storage/FileDB.log -source-path=.

import (
	"fmt"
	"log"
	"net/http"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/routes"
)

func main() {
	// config.SetFlag()
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
