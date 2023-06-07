package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/grishagavrin/link-shortener/internal/utils/db"
	"go.uber.org/zap"
)

func main() {
	// Context with cancel func
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

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

	go func() {
		l.Fatal("app error exit", zap.Error(srv.ListenAndServe()))
	}()
	l.Info("The service is ready to listen and serve.")

	// Add context for Graceful shutdown
	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		l.Info("Got SIGINT...")
	case syscall.SIGTERM:
		l.Info("Got SIGTERM...")
	}

	// database close
	conn, err := db.Instance()
	if err == nil {
		l.Info("Closing connect to db")
		conn.Close()
	}

	l.Info("Closing connect to db success")
}
