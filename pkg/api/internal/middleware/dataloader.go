package middleware

import (
	"context"
	"fmt"
	"github.com/graph-gophers/dataloader/v7"
	"github.com/jollyboss123/tcg_my-server/pkg/api/internal/model"
	"github.com/jollyboss123/tcg_my-server/pkg/source"
	"net/http"
	"strings"
	"time"
)

type contextKey string

const detailLoaderCtxKey = contextKey("detailLoader")

func NewDataLoader(wait time.Duration, maxBatch int, ds source.DetailService) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			ctx = context.WithValue(ctx, detailLoaderCtxKey, newBatchLoader(detailBatchFn(ds), wait, maxBatch))

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func DetailLoaderFromContext(ctx context.Context) (*dataloader.Loader[string, *model.DetailInfo], error) {
	return dataLoaderFromContext[string, *model.DetailInfo](ctx, detailLoaderCtxKey)
}

func newBatchLoader[K comparable, V any](batchFn func(context.Context, []K) []*dataloader.Result[V], wait time.Duration, maxBatch int) *dataloader.Loader[K, V] {
	return dataloader.NewBatchedLoader(batchFn, dataloader.WithWait[K, V](wait), dataloader.WithInputCapacity[K, V](maxBatch))
}

func dataLoaderFromContext[K comparable, T any](ctx context.Context, contextKey contextKey) (*dataloader.Loader[K, T], error) {
	dataLoader, ok := ctx.Value(contextKey).(*dataloader.Loader[K, T])
	if !ok {
		var nodeType T
		return nil, fmt.Errorf("%T data loader not found", nodeType)
	}
	return dataLoader, nil
}

func detailBatchFn(ds source.DetailService) func(context.Context, []string) []*dataloader.Result[*model.DetailInfo] {
	return func(ctx context.Context, keys []string) []*dataloader.Result[*model.DetailInfo] {
		results := make([]*dataloader.Result[*model.DetailInfo], len(keys))
		for i, key := range keys {
			parts := strings.Split(key, "|")
			code := parts[0]
			game := parts[1]
			d, err := ds.Fetch(ctx, code, game)
			if err != nil {
				results[i] = &dataloader.Result[*model.DetailInfo]{
					Error: err,
				}
				continue
			}
			results[i] = &dataloader.Result[*model.DetailInfo]{
				Data: model.ToDetailInfo(d),
			}
		}
		return results
	}
}
