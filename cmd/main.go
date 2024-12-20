package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mpx/config"
	"mpx/internal/dependencies"
	"mpx/internal/transport"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-pkgz/routegroup"
	"golang.org/x/sync/errgroup"
)

const (
	// readHeaderTimeout timeout for HTTP header reading
	readHeaderTimeout = 5 * time.Second
	// maximum number of simultaneous requests
	maxConcurrentRequests = 100
)

// Semaphore to limit the number of simultaneous requests
var semaphore = make(chan struct{}, maxConcurrentRequests)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err = run(ctx, cfg); err != nil {
		log.Printf("error running the service: %s", err)
	}
}

func run(ctx context.Context, cfg *config.Config) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	deps, err := dependencies.NewContainer(ctx, cfg)
	if err != nil {
		return fmt.Errorf("error creating dependencies: %w", err)
	}

	ctx = deps.Logger.WithContext(ctx)

	serverHandler := getHandler(deps)

	server := &http.Server{
		Addr:    deps.Config.Server.Address,
		Handler: serverHandler,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		ReadHeaderTimeout: readHeaderTimeout,
	}

	eg := new(errgroup.Group)

	eg.Go(func() error {
		deps.Logger.Info().Str("addr", server.Addr).Msg("start listening")
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("error starting server: %w", err)
		}

		return nil
	})

	// Graceful shutdown
	eg.Go(func() error {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

		select {
		case <-sigCh:
			deps.Logger.Info().Msg("Received shutdown signal")
			shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
			defer shutdownCancel()

			if err = server.Shutdown(shutdownCtx); err != nil {
				return fmt.Errorf("graceful shutdown failed: %w", err)
			}
			deps.Logger.Info().Msg("Server gracefully stopped")
		}

		return nil
	})

	err = eg.Wait()
	if err != nil {
		return fmt.Errorf("error starting server: %w", err)
	}

	return nil
}

func getHandler(deps *dependencies.Container) http.Handler {
	mux := http.NewServeMux()

	r := routegroup.New(mux)

	h := deps.Handlers

	r.Handle("POST /", transport.WithErrorHandling(semaphore, h.Handler.Send))

	return r
}
