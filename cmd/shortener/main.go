package main

import (
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
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
		l.Fatal("Config instance error: ", zap.Error(err))
	}

	//DB Instance
	dbi, err := db.Instance(l)
	if errors.Is(err, errs.ErrDatabaseNotAvaliable) {
		l.Info("DB error", zap.Error(err))
	}

	srv := &http.Server{
		Addr:    srvAddr,
		Handler: routes.ServiceRouter(l),
	}

	go func() {
		l.Fatal("App error exit", zap.Error(srv.ListenAndServe()))
	}()
	l.Info("The server is ready")

	// Add context for Graceful shutdown
	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		l.Info("Got SIGINT...")
	case syscall.SIGTERM:
		l.Info("Got SIGTERM...")
	}

	// Database close
	if err == nil {
		l.Info("Closing connect to db")
		dbi.Close()
	}

	l.Info("Closing connect to db success")
}
