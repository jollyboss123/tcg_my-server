package api

import (
	"context"
	"github.com/go-chi/chi/v5"
	"github.com/jollyboss123/tcg_my-server/config"
	redisLib "github.com/jollyboss123/tcg_my-server/pkg/redis"
	"github.com/redis/go-redis/v9"
	"github.com/rs/cors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server struct {
	Version string
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
			log.Fatalln(err)
		}
	}
	return s
}

func WithVersion(version string) Options {
	return func(opts *Server) error {
		log.Printf("Starting API version: %s\n", version)
		opts.Version = version
		return nil
	}
}

func defaultServer() *Server {
	return &Server{
		cfg: config.New(),
	}
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

	s.cache = redisLib.New(s.cfg.Cache)
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
	log.Printf("Serving at %s:%s\n", s.cfg.Api.Host, s.cfg.Api.Port)
	err := s.httpServer.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func gracefulShutdown(ctx context.Context, s *Server) error {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit

	log.Println("Shutting down...")

	ctx, shutdown := context.WithTimeout(ctx, s.cfg.Api.GracefulTimeout*time.Second)
	defer shutdown()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		log.Println(err)
	}
	s.closeResources(ctx)

	return nil
}

func (s *Server) closeResources(ctx context.Context) {
	s.cache.Shutdown(ctx)
}
