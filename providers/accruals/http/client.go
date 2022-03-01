package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/zhupanovdm/gophermart/config"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/server"
	"github.com/zhupanovdm/gophermart/providers/accruals"
	"github.com/zhupanovdm/gophermart/providers/accruals/model"
)

const (
	RetryAfterHeader = "Retry-After"

	accrualsHTTPClientName = "Accruals HTTP Client"
)

var _ accruals.Accruals = (*client)(nil)

type client struct {
	url *url.URL
	http.Client
}

func (c client) Get(ctx context.Context, number string) (*model.AccrualResponse, error) {
	ctx, logger := logging.ServiceLogger(ctx, accrualsHTTPClientName)
	logger.Info().Msg("querying accruals service")

	accrualsURL := *c.url
	accrualsURL.Path = path.Join("/api/orders", number)
	logger.Info().Msgf("calling: %s", accrualsURL.String())

	req, err := http.NewRequestWithContext(ctx, "GET", accrualsURL.String(), nil)
	if err != nil {
		logger.Err(err).Msg("failed to prepare request")
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		logger.Err(err).Msg("failed to call service")
		return nil, err
	}

	body := resp.Body
	//goland:noinspection GoUnhandledErrorResult
	defer body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrual model.AccrualResponse
		if err := json.NewDecoder(body).Decode(&accrual); err != nil {
			logger.Err(err).Msg("failed to decode json response")
			return nil, err
		}
		logger.Trace().Msg("got accrual result")
		return &accrual, nil
	case http.StatusNoContent:
		logger.Warn().Msg("no content")
		return nil, nil
	case http.StatusTooManyRequests:
		err := accruals.NewError(accruals.ErrTooManyRequest, http.StatusText(resp.StatusCode))
		times, _ := strconv.Atoi(resp.Header.Get(RetryAfterHeader))
		err.RetryAfter = time.Duration(times) * time.Second
		logger.Warn().Msgf("too many requests: should wait for %v", err.RetryAfter)
		return nil, err
	}

	err = fmt.Errorf("response code %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	logger.Err(err).Msg("failed to call accrual service")
	return nil, err
}

func New(cfg *config.Config) (accruals.Accruals, error) {
	serviceURL, err := server.ParseURL(cfg.AccrualSystemAddress)
	if err != nil {
		return nil, fmt.Errorf("error retrieving accruals service URL: %w", err)
	}
	return &client{
		url: serviceURL,
	}, nil
}

func Factory(cfg *config.Config) accruals.ClientFactory {
	return func() (accruals.Accruals, error) {
		return New(cfg)
	}
}
