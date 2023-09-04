// Main package for initial initialization of the application
package main

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/logger"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/grishagavrin/link-shortener/internal/storage"
	istorage "github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"go.uber.org/zap"
)

func main() {
	// Seed install for math/rand
	rand.Seed(time.Now().UnixNano())

	// Context with cancel func
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Logger instance
	l, err := logger.Instance()
	if err != nil {
		log.Fatal("fatal logger:", zap.Error(err))
	}

	// Config instance
	cfg, err := config.Instance()
	if errors.Is(err, errs.ErrENVLoading) {
		log.Fatal(errs.ErrConfigInstance, zap.Error(err))
	}

	// Get server address
	srvAddr, err := cfg.GetCfgValue(config.ServerAddress)
	if errors.Is(err, errs.ErrUnknownEnvOrFlag) {
		l.Fatal("fatal get config value: ", zap.Error(err))
	}

	// Batch channel for batch delete
	chBatch := make(chan istorage.BatchDelete)

	// Storage instance allocate logger and batch channel
	stor, err := storage.Instance(l, chBatch)
	if err != nil {
		l.Fatal("fatal storage init", zap.Error(err))
	}

	srv := &http.Server{
		Addr:    srvAddr,
		Handler: routes.ServiceRouter(stor.Repository, l, chBatch),
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

	close(chBatch)

	// Database close
	if err == nil && stor.SQLDB != nil {
		l.Info("closing connect to db")
		stor.SQLDB.Close()
	}
}
