package api

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/jollyboss123/tcg_my-server/config"
	redisLib "github.com/jollyboss123/tcg_my-server/pkg/redis"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	Version string
	log     *slog.Logger
	cfg     *config.Config

	cache *redis.Client

	cors       *cors.Cors
	router     *chi.Mux
	httpServer *http.Server
}

type Options func(opts *Server) error

func NewServer(opts ...Options) *Server {
	s := defaultServer()

	for _, opt := range opts {
		err := opt(s)
		if err != nil {
			s.log.Error("build server", slog.String("error", err.Error()))
		}
	}
	return s
}

func WithVersion(version string) Options {
	return func(opts *Server) error {
		child := opts.log.With(
			slog.String("version", version))
		opts.log = child
		opts.Version = version
		return nil
	}
}

func defaultServer() *Server {
	logger := defaultLogger()

	return &Server{
		log: logger,
		cfg: config.New(logger),
	}
}

func defaultLogger() *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, nil)
	logger := slog.New(handler)
	return logger.With(
		slog.String("app", "tcg.my"))
}

func (s *Server) Init() {
	s.setCors()
	s.newRedis()
	s.newRouter()
	s.setGlobalMiddleware()
	s.InitRouter()
}

func (s *Server) setCors() {
	s.cors = cors.New(cors.Options{
		AllowedOrigins:      s.cfg.Cors.AllowedOrigins,
		AllowedMethods:      []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:      []string{"*"},
		AllowCredentials:    s.cfg.Cors.AllowCredentials,
		AllowPrivateNetwork: s.cfg.Cors.AllowPrivateNetwork,
	})
}

func (s *Server) setGlobalMiddleware() {
	s.router.Use(s.cors.Handler)
}

func (s *Server) newRedis() {
	if !s.cfg.Cache.Enable {
		return
	}

	client, err := redisLib.New(s.cfg.Cache, s.log)
	if err != nil {
		return
	}
	s.cache = client
}

func (s *Server) newRouter() {
	s.router = chi.NewRouter()
}

func (s *Server) Run() {
	s.httpServer = &http.Server{
		Addr:              s.cfg.Api.Host + ":" + s.cfg.Api.Port,
		Handler:           s.router,
		ReadHeaderTimeout: s.cfg.Api.ReadHeaderTimeout,
	}

	go func() {
		start(s)
	}()

	_ = gracefulShutdown(context.Background(), s)
}

func start(s *Server) {
	s.log.LogAttrs(context.Background(),
		slog.LevelInfo,
		"server started",
		slog.String("address", fmt.Sprintf("%s:%s", s.cfg.Api.Host, s.cfg.Api.Port)))
	err := s.httpServer.ListenAndServe()
	if err != nil {
		s.log.Error("listen and serve", slog.String("error", err.Error()))
	}
}

func gracefulShutdown(ctx context.Context, s *Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	s.log.Log(ctx, slog.LevelInfo, "shutting down")

	ctx, shutdown := context.WithTimeout(ctx, s.cfg.Api.GracefulTimeout*time.Second)
	defer shutdown()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		s.log.Warn("server shutdown", slog.String("error", err.Error()))
	}
	s.closeResources(ctx)

	return nil
}

func (s *Server) closeResources(ctx context.Context) {
	s.cache.Shutdown(ctx)
}
