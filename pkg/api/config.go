package api

import (
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/query"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
)

//go:generate go run github.com/99designs/gqlgen --config ../../gqlgen.yml generate
func newConfig(scrape source.ScrapeService) resolver.Config {
	return resolver.Config{
		Resolvers: &resolverRoot{
			queryResolver: query.NewQueryResolver(scrape),
		},
	}
}
