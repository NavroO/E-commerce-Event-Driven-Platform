package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NavroO/ecommerce-platform/internal/order"
	"github.com/NavroO/ecommerce-platform/pkg/config"
	"github.com/NavroO/ecommerce-platform/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	mongopkg "github.com/NavroO/ecommerce-platform/pkg/mongo"
	rabbitpkg "github.com/NavroO/ecommerce-platform/pkg/rabbit"
)

func main() {
	cfg := config.FromEnv()
	log := logger.New(cfg.LogLevel).With().Str("svc", cfg.ServiceName).Logger()

	ctx := context.Background()

	mc, err := mongopkg.Connect(ctx, cfg.MongoURI, log)
	must(err, log)
	db := mc.Database(cfg.MongoDB)
	must(mongopkg.EnsureIndexes(ctx, db, log), log)

	ordersCol := db.Collection("orders")
	repo := order.NewRepository(ordersCol)

	pub, err := rabbitpkg.ConnectWithRetry(ctx, cfg.RabbitURI, cfg.EventsExchange, log, 10, 3*time.Second)
	must(err, log)
	defer pub.Close()

	svc := order.NewService(repo, pub, log)

	mux := http.NewServeMux()
	mux.Handle("/", order.Routes(svc))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("/metrics", promhttp.Handler())

	srv := &http.Server{
		Addr:         cfg.HTTPAddr,
		Handler:      loggingMiddleware(log)(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info().Str("addr", cfg.HTTPAddr).Msg("order-service listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("server failed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctxShutdown)
	_ = mc.Disconnect(ctxShutdown)
}

func must(err error, log zerolog.Logger) {
	if err != nil {
		log.Fatal().Err(err).Msg("fatal")
	}
}

func loggingMiddleware(log zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Dur("dur", time.Since(start)).
				Msg("http")
		})
	}
}
