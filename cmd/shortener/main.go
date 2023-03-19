package main

// macOS: nodemon --watch ../../ --exec go run main.go --signal SIGTERM
// wnd: nodemon --watch ../../ --exec go run main.go --signal SIGKILL
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
	srvAddr, err := cfg.GetCfgValue(config.ServerAddress)
	if err != nil {
		log.Fatal(err.Error())
	}

	srv := &http.Server{
		Addr:    srvAddr,
		Handler: routes.ServiceRouter(),
	}

	fmt.Printf("Server started on %s\n", srvAddr)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
