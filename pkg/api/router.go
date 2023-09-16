package api

import (
	gqlhandler "github.com/99designs/gqlgen/graphql/handler"
	gqlplayground "github.com/99designs/gqlgen/graphql/playground"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/go-chi/chi/v5"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
)

func (s *Server) InitRouter() {
	executableSchemeConfig := newConfig(source.NewYYT())

	gqlHandler := gqlhandler.New(resolver.NewExecutableSchema(executableSchemeConfig))
	gqlHandler.AddTransport(transport.GET{})
	gqlHandler.AddTransport(transport.POST{})
	gqlHandler.AddTransport(transport.MultipartForm{})
	gqlHandler.AddTransport(transport.UrlEncodedForm{})
	gqlHandler.AddTransport(transport.SSE{})
	gqlHandler.AddTransport(transport.Options{})

	gqlHandler.Use(extension.Introspection{})
	gqlHandler.Use(extension.FixedComplexityLimit(500))

	gqlHandler.Use(extension.AutomaticPersistedQuery{Cache: s.cache})

	s.router.Group(func(r chi.Router) {
		handlerFunc := gqlplayground.Handler("GraphiQL Playground", "/query")
		r.Handle("/graphiql", handlerFunc)
		r.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {})
	})

	s.router.Group(func(r chi.Router) {
		r.Handle("/query", gqlHandler)
		r.Handle("/api/query", gqlHandler)
	})
}
