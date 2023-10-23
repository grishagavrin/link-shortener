// Main package for initial initialization of the application
package main

import (
	"errors"
	"fmt"
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
	istorage "github.com/grishagavrin/link-shortener/internal/storage/iStorage"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
)

// @Title Link Shortener API
// @Description Link shortener service
// @Version 1.0

// @Contact.email grigorygavrin@gmail.com

// @BasePath /
// @Host 127.0.0.1:8080

// Global variables
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	// Print build info
	printBuildInfo()
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
	// HTTP server
	if cfg.EnableHTTPS == "" {
		srv := &http.Server{
			Addr:    srvAddr,
			Handler: routes.ServiceRouter(stor.Repository, l, chBatch),
		}

		go func() {
			l.Info("Start HTTP server")
			err := srv.ListenAndServe()
			if err != nil {
				l.Info("app error exit", zap.Error(err))
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}
		}()
	} else {
		// HTTPS server
		manager := &autocert.Manager{
			Cache:      autocert.DirCache("cache-dir"),
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(srvAddr),
		}

		srv := &http.Server{
			Addr:      ":443",
			Handler:   routes.ServiceRouter(stor.Repository, l, chBatch),
			TLSConfig: manager.TLSConfig(),
		}

		go func() {
			l.Info("Start HTTPS server")
			err := srv.ListenAndServeTLS("server.crt", "server.key")
			if err != nil {
				l.Info("app error exit", zap.Error(err))
				syscall.Kill(syscall.Getpid(), syscall.SIGINT)
			}
		}()

	}

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

// printBuildInfo print info about package
func printBuildInfo() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}

	if buildDate == "" {
		buildDate = "N/A"
	}

	if buildCommit == "" {
		buildCommit = "N/A"
	}

	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}
