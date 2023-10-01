package proxy

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"log/slog"
	"sync/atomic"
)

type service struct {
	proxies []*apigateway.RestApi
	session *session.Session
	index   uint32
	logger  *slog.Logger
}

type Service interface {
	FetchProxyURL() (string, error)
}

func NewService(logger *slog.Logger) Service {
	child := logger.With(slog.String("api", "proxy"))
	return &service{
		logger: child,
	}
}

const region = "ap-southeast-1" //hardcoded to southeast for now
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

// put in cache
func (s *service) fetchProxies() ([]*apigateway.RestApi, error) {
	emptyResult := make([]*apigateway.RestApi, 0)
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
	if err != nil {
		s.logger.Error("create aws session", slog.String("error", err.Error()))
		return emptyResult, err
	}

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

	return result.Items, nil
}

func (s *service) RoundRobinProxy(targetURL string) (string, error) {
	if len(targetURL) == 0 {
		return "", ErrEmptyTargetURL
	}

	index := atomic.AddUint32(&s.index, 1) - 1

	if index%uint32(len(s.proxies)+1) == 0 {
		return targetURL, nil
	}

	proxy := s.proxies[index%uint32(len(s.proxies))]

	return fmt.Sprintf("https://%s.execute-api.%s.amazonaws.com/%s/%s",
		*proxy.Id,
		&s.session.Config.Region,
		"prod", targetURL), nil
}
