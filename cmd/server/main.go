package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"ads-txt-service/internal/cache"
	"ads-txt-service/internal/config"
	"ads-txt-service/internal/fetcher"
	"ads-txt-service/internal/handler"
	"ads-txt-service/internal/logger"
	"ads-txt-service/internal/parser"
)

type Application struct {
	cfg          *config.Config
	log          *logger.Logger
	cache        cache.Cache
	httpServer   *http.Server
	shutdownWait sync.WaitGroup
}

func NewApplication() (*Application, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	if err := logger.Init(cfg.LogLevel); err != nil {
		return nil, fmt.Errorf("init logger: %w", err)
	}

	log := logger.L()

	cacheBackend, err := cache.InitCache(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to init cache: %w", err)
	}

	adsCache := cache.NewAdsCache(cacheBackend)

	ft := fetcher.NewFetcher(cfg.HttpClientTO)

	pr := parser.NewParser()

	srv := handler.NewServer(cfg, adsCache, log, ft, pr)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      srv.Router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	return &Application{
		cfg:        cfg,
		log:        log,
		cache:      cacheBackend,
		httpServer: httpServer,
	}, nil
}

func (a *Application) Run(ctx context.Context) error {
	a.shutdownWait.Add(1)
	go func() {
		defer a.shutdownWait.Done()
		a.log.Infof("Server starting on %s", a.httpServer.Addr)
		if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-ctx.Done()
	a.log.Info("Shutdown signal received")
	return a.Shutdown()
}

func (a *Application) Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	if err := a.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown HTTP: %w", err)
	}

	a.shutdownWait.Wait()

	if closer, ok := a.cache.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			return fmt.Errorf("close cache: %w", err)
		}
	}

	a.log.Infow("Server gracefully stopped", "duration", time.Since(start))
	return nil
}

func main() {
	app, err := NewApplication()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Application init error: %v\n", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		app.log.Errorf("Application error: %v", err)
		os.Exit(1)
	}
}