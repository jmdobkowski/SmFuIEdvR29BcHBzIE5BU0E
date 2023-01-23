package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jmdobkowski/SmFuIEdvR29BcHBzIE5BU0E/url-collector/config"
	"github.com/jmdobkowski/SmFuIEdvR29BcHBzIE5BU0E/url-collector/internal"
	"github.com/jmdobkowski/SmFuIEdvR29BcHBzIE5BU0E/url-collector/internal/providers"
)

func main() {
	cfg := config.LoadFromEnv()

	provider := providers.NewAPODProvider(cfg.ApiKey, cfg.ConcurrentRequests)
	server := internal.Server{Provider: provider}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	hs := http.Server{
		Addr:        ":" + strconv.Itoa(cfg.Port),
		Handler:     server,
		BaseContext: func(l net.Listener) context.Context { return ctx },
	}

	go func() {
		<-ctx.Done()
		log.Printf("shutting down server...")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		hs.Shutdown(shutdownContext)
	}()

	log.Printf("starting server on port %d...", cfg.Port)
	log.Fatal(hs.ListenAndServe())
}
