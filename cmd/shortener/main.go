package main

import (
	"errors"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"go.uber.org/zap"
)

func main() {
	// Context with cancel func
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Logger instance
	l, err := logger.Instance()
	if errors.Is(err, errs.ErrInitLogger) {
		log.Fatal("fatal logger instance: ", zap.Error(err))
	}

	srvAddr, err := config.Instance().GetCfgValue(config.ServerAddress)
	if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
		l.Fatal("fatal get config value: ", zap.Error(err))
	}

	// Storage instance
	stor, dbi, err := storage.Instance(l)
	if err != nil {
		l.Fatal("fatal storage init", zap.Error(err))
	}

	srv := &http.Server{
		Addr:    srvAddr,
		Handler: routes.ServiceRouter(stor, l),
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
	if err == nil && dbi != nil {
		l.Info("closing connect to db")
		dbi.Close()
	}
}
