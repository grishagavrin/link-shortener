// Main package for initial initialization of the application
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"github.com/grishagavrin/link-shortener/internal/config"
	"github.com/grishagavrin/link-shortener/internal/errs"
	"github.com/grishagavrin/link-shortener/internal/handlers"
	handlersgrpc "github.com/grishagavrin/link-shortener/internal/handlersGPRC"
	"github.com/grishagavrin/link-shortener/internal/logger"
	ls "github.com/grishagavrin/link-shortener/internal/proto"
	"github.com/grishagavrin/link-shortener/internal/routes"
	"github.com/grishagavrin/link-shortener/internal/storage"
	"github.com/grishagavrin/link-shortener/internal/storage/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/grpc"
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

	// Logger instance
	l, err := logger.Instance()
	if err != nil {
		log.Fatal("fatal logger:", zap.Error(err))
	}

	// Init context
	ctx, cancel := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)
	defer cancel()

	// Batch channel for batch delete
	chBatch := make(chan models.BatchDelete)

	// Storage instance allocate logger and batch channel
	stor, err := storage.Instance(l, chBatch)
	if err != nil {
		l.Fatal("fatal storage init", zap.Error(err))
	}

	// Handlers REST
	h := handlers.New(stor.Repository, l)

	//grpc
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("cannot create listener: %s", err)
	}

	serverRegistrar := grpc.NewServer()
	hGRPC := handlersgrpc.New(stor.Repository, l)
	// service := &grpc_handlers.GRPCHandlers{}
	ls.RegisterApiServiceServer(serverRegistrar, hGRPC)

	go func() {
		fmt.Println("GRPC Started on port 50051")
		err = serverRegistrar.Serve(lis)
		if err != nil {
			log.Fatalf("impossible to serve: %s", err)
		}
	}()

	///

	// Routing app
	r := routes.NewRouterFacade(h, l, chBatch)

	// Start server
	startServer(ctx, r, l, stor, chBatch)
}

// start server function
func startServer(
	ctx context.Context,
	r *routes.RouterFacade,
	l *zap.Logger,
	stor *storage.InstanceStruct,
	chBatch chan models.BatchDelete,
) {

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

	if cfg.EnableHTTPS == "" {
		// Start func for HTTP server
		srv := startHTTPServer(srvAddr, r.HTTPRoute.Route, l)
		releaseResources(ctx, l, stor, chBatch, srv)
	} else {
		// Start func for HTTPS server
		srv := startHTTPSServer(srvAddr, r.HTTPRoute.Route, l)
		releaseResources(ctx, l, stor, chBatch, srv)
	}
}

// Release of resources app
func releaseResources(ctx context.Context,
	l *zap.Logger,
	stor *storage.InstanceStruct,
	chBatch chan models.BatchDelete,
	srv *http.Server,
) {
	<-ctx.Done()
	if ctx.Err() != nil {
		fmt.Printf("Error:%v\n", ctx.Err())
	}

	l.Info("The service is shutting down...")
	if stor.SQLDB != nil {
		l.Info("closing connect to db")
		stor.SQLDB.Close()
	}

	// Close channel of batch delete
	close(chBatch)

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		l.Info("app error exit", zap.Error(err))
	}
}

// Start HTTP server func
func startHTTPServer(
	srvAddr string,
	h http.Handler,
	l *zap.Logger,
) *http.Server {
	srv := &http.Server{
		Addr:    srvAddr,
		Handler: h,
	}
	go func() {
		l.Info("Start HTTP server")
		err := srv.ListenAndServe()
		if err != nil {
			l.Info("app error exit", zap.Error(err))
		}
	}()

	return srv
}

// Start HTTPS server func
func startHTTPSServer(
	srvAddr string,
	h http.Handler,
	l *zap.Logger,
) *http.Server {
	manager := &autocert.Manager{
		Cache:      autocert.DirCache("cache-dir"),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(srvAddr),
	}

	srv := &http.Server{
		Addr:      ":443",
		Handler:   h,
		TLSConfig: manager.TLSConfig(),
	}

	go func() {
		l.Info("Start HTTPS server")
		err := srv.ListenAndServeTLS("server.crt", "server.key")
		if err != nil {
			l.Info("app error exit", zap.Error(err))
		}
	}()

	return srv
}

// Print build info print info about package
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
