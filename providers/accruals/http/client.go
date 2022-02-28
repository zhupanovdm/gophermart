package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/zhupanovdm/gophermart/providers/accruals"
	"github.com/zhupanovdm/gophermart/providers/accruals/model"
)

const RetryAfterHeader = "Retry-After"

var _ accruals.Accruals = (*client)(nil)

type client struct {
	baseUrl string
	http.Client
}

func (c client) Get(ctx context.Context, number string) (*model.AccrualResponse, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/api/orders/%s", c.baseUrl, number), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	body := resp.Body
	defer body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var accrual model.AccrualResponse
		if err := json.NewDecoder(body).Decode(&accrual); err != nil {
			return nil, err
		}
		return &accrual, nil
	case http.StatusTooManyRequests:
		err := accruals.NewError(accruals.ErrTooManyRequest, http.StatusText(resp.StatusCode))
		times, _ := strconv.Atoi(resp.Header.Get(RetryAfterHeader))
		err.RetryAfter = time.Duration(times) * time.Second
		return nil, err
	}
	return nil, fmt.Errorf("response code %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
}

func New(baseUrl string) accruals.Accruals {
	return &client{
		baseUrl: baseUrl,
	}
}
