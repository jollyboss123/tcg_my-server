package api

import "github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"

type resolverRoot struct {
	queryResolver resolver.QueryResolver
	cardResolver  resolver.CardResolver
}

func (r *resolverRoot) Query() resolver.QueryResolver {
	return r.queryResolver
}

func (r *resolverRoot) Card() resolver.CardResolver {
	return r.cardResolver
}
