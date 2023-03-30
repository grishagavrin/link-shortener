package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/routes"
)

func main() {
	srvAddr, err := config.Instance().GetCfgValue(config.ServerAddress)
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
