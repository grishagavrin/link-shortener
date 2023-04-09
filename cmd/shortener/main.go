package main

import (
	"log"
	"net/http"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"go.uber.org/zap"
)

func main() {
	l, err := logger.Instance()
	if err != nil {
		log.Fatal(err)
	}
	srvAddr, err := config.Instance().GetCfgValue(config.ServerAddress)
	if err != nil {
		l.Fatal("app error exit", zap.Error(err))
	}

	srv := &http.Server{
		Addr:    srvAddr,
		Handler: routes.ServiceRouter(),
	}

	l.Info("Start server address: " + srvAddr)
	err = srv.ListenAndServe()
	if err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
