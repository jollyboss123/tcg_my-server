package api

import (
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/query"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"
	"github.com/jollyboss123/tcg_my-server/pkg/currency"
	"github.com/jollyboss123/tcg_my-server/pkg/rate"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
)

//go:generate go run github.com/99designs/gqlgen --config ../../gqlgen.yml generate
func newConfig(scrape source.ScrapeService, currency currency.Service, rate rate.Service) resolver.Config {
	return resolver.Config{
		Resolvers: &resolverRoot{
			queryResolver: query.NewQueryResolver(scrape, currency, rate),
		},
	}
}
