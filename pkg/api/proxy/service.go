package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/jollyboss123/tcg_my-server/config"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"sync/atomic"
)

type service struct {
	proxies []*apigateway.RestApi
	region  string
	index   uint32
	cache   *redis.Client
	cfg     *config.Config
	logger  *slog.Logger
}

type Service interface {
	FetchProxyURL() (string, error)
	RoundRobinProxy(ctx context.Context, targetURL string) (string, error)
}

func NewService(logger *slog.Logger, cache *redis.Client, cfg *config.Config) Service {
	child := logger.With(slog.String("api", "proxy"))
	return &service{
		logger: child,
		cache:  cache,
		cfg:    cfg,
	}
}

const (
	region = "ap-southeast-1" //hardcoded to southeast for now
	key    = "proxy"
)

var (
	ErrNoAPIGatewayInstances = errors.New("no api gateway instances found")
	ErrEmptyTargetURL        = errors.New("target url given is empty")
)

func (s *service) FetchProxyURL() (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		s.logger.Error("create aws session", slog.String("error", err.Error()))
		return "", err
	}

	svc := apigateway.New(sess)

	result, err := svc.GetRestApis(nil)
	if err != nil {
		s.logger.Error("fetch api gateway list", slog.String("error", err.Error()))
		return "", err
	}

	if len(result.Items) == 0 {
		s.logger.Error("checking api gateway list", slog.String("error", ErrNoAPIGatewayInstances.Error()))
		return "", ErrNoAPIGatewayInstances
	}

	proxy := result.Items[0]

	return fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s",
		*proxy.Id,
		*sess.Config.Region,
		"prod"), nil
}

func (s *service) fetchProxies(ctx context.Context) ([]*apigateway.RestApi, error) {
	emptyResult := make([]*apigateway.RestApi, 0)
	var proxies []*apigateway.RestApi
	val, err := s.cache.Get(ctx, key).Result()
	if err != nil {
		s.logger.Warn("get proxy cache", slog.String("error", err.Error()))
	}
	err = json.Unmarshal([]byte(val), &proxies)
	if err == nil {
		s.logger.Info("cache hit", slog.String("key", key))
		return proxies, nil
	}
	s.logger.Info("cache miss", slog.String("key", key))

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		s.logger.Error("create aws session", slog.String("error", err.Error()))
		return emptyResult, err
	}

	s.region = *sess.Config.Region
	svc := apigateway.New(sess)

	result, err := svc.GetRestApis(nil)
	if err != nil {
		s.logger.Error("fetch api gateway list", slog.String("error", err.Error()))
		return emptyResult, err
	}

	if len(result.Items) == 0 {
		s.logger.Error("checking api gateway list", slog.String("error", ErrNoAPIGatewayInstances.Error()))
		return emptyResult, err
	}

	data, err := json.Marshal(result.Items)
	if err == nil {
		err = s.cache.Set(ctx, key, data, s.cfg.Cache.CacheTime).Err()
		if err != nil {
			s.logger.Warn("set proxy cache", slog.String("error", err.Error()))
		}
	}

	return result.Items, nil
}

func (s *service) RoundRobinProxy(ctx context.Context, targetURL string) (string, error) {
	if !s.cfg.Api.ProxyEnabled {
		s.logger.Info("proxy disabled")
		return targetURL, nil
	}

	if len(targetURL) == 0 {
		s.logger.Error("check target url", slog.String("error", ErrEmptyTargetURL.Error()))
		return "", ErrEmptyTargetURL
	}

	proxies, err := s.fetchProxies(ctx)
	if err != nil {
		s.logger.Error("fetch proxies", slog.String("error", err.Error()), slog.String("target", targetURL))
		return "", err
	}
	s.proxies = proxies
	index := atomic.AddUint32(&s.index, 1) - 1

	if index%uint32(len(s.proxies)+1) == 0 {
		return targetURL, nil
	}

	proxy := s.proxies[index%uint32(len(s.proxies))]

	return fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s/%s",
		*proxy.Id,
		s.region,
		"prod", targetURL), nil
}
