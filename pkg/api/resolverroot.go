package api

import "github.com/jollyboss123/tcg_my-server/pkg/api/internal/resolver"

type resolverRoot struct {
	queryResolver resolver.QueryResolver
}

func (r *resolverRoot) Query() resolver.QueryResolver {
	return r.queryResolver
}
