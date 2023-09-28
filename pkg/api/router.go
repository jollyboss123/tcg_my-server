package api

import (
	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	gqlplayground "github.com/99designs/gqlgen/graphql/playground"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/middleware"
	"github.com/jollyboss123/tcg_my-server/pkg/currency"
	"github.com/jollyboss123/tcg_my-server/pkg/game"
	"github.com/jollyboss123/tcg_my-server/pkg/rate"
	"github.com/jollyboss123/tcg_my-server/pkg/source/detail"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/go-chi/chi/v5"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
)

const (
	dataLoaderWait     = 250 * time.Microsecond
	dataLoaderMaxBatch = 1000
)

func (s *Server) InitRouter() {
	executableSchemeConfig := newConfig(
		s.currencyService(),
		s.scrapeService(),
		s.rateService(),
		s.gameService(),
		s.detailService(),
	)

	gqlHandler := gqlhandler.New(resolver.NewExecutableSchema(executableSchemeConfig))
	gqlHandler.AddTransport(transport.GET{})
	gqlHandler.AddTransport(transport.POST{})
	gqlHandler.AddTransport(transport.MultipartForm{})
	gqlHandler.AddTransport(transport.UrlEncodedForm{})
	gqlHandler.AddTransport(transport.SSE{})
	gqlHandler.AddTransport(transport.Options{})

	gqlHandler.Use(extension.Introspection{})
	gqlHandler.Use(extension.FixedComplexityLimit(500))

	gqlHandler.Use(extension.AutomaticPersistedQuery{Cache: lru.New(1000)})

	s.router.Group(func(r chi.Router) {
		handlerFunc := gqlplayground.Handler("GraphiQL Playground", "/query")
		r.Handle("/graphiql", handlerFunc)
		r.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {})
	})

	s.router.Group(func(r chi.Router) {
		r.Use(middleware.NewDataLoader(
			dataLoaderWait,
			dataLoaderMaxBatch,
			s.detailService(),
		))

		r.Handle("/query", gqlHandler)
		r.Handle("/api/query", gqlHandler)
	})
}

func (s *Server) scrapeService() source.ScrapeService {
	return source.NewCachedScrapeService(
		s.cache,
		s.cfg,
		s.log,
		s.gameService(),
		source.NewYYT(s.log, s.currencyService(), s.gameService()),
		//source.NewBigWeb(s.log), //disabled bigweb for now
	)
}

func (s *Server) currencyService() currency.Service {
	return currency.NewService(s.log)
}

func (s *Server) rateService() rate.Service {
	return rate.NewCachedExchangeRate(
		rate.NewService(s.log, s.cfg, s.currencyService()),
		s.cache,
		s.cfg,
		s.log,
		s.currencyService(),
	)
}

func (s *Server) gameService() game.Service {
	return game.NewService(s.log)
}

func (s *Server) detailService() source.DetailService {
	//services := map[string]source.DetailService{
	//	game.YGO: detail.NewYGO(s.log, s.gameService()),
	//}
	//
	//return detail.NewCachedDetailService(
	//	s.cache,
	//	s.cfg,
	//	s.log,
	//	services,
	//	s.gameService(),
	//)
	return detail.NewYGO(s.log, s.gameService())
}
