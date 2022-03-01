package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zhupanovdm/gophermart/model/balance"
	"github.com/zhupanovdm/gophermart/pkg/errors"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/server"
	"github.com/zhupanovdm/gophermart/service"
)

const balanceHandlerName = "Balance Handler"

type balanceHandler struct {
	service.Balance
}

func (h *balanceHandler) Get(resp http.ResponseWriter, req *http.Request) {
	//goland:noinspection GoUnhandledErrorResult
	defer req.Body.Close()

	ctx, logger := logging.ServiceLogger(req.Context(), balanceHandlerName)
	logger.Info().Msg("handling client balance query")

	currentBalance, err := h.Balance.Get(ctx, AuthorizedUserID(ctx))
	if err != nil {
		logger.Err(err).Msg("failed to query balance")
		server.Error(resp, http.StatusInternalServerError, nil)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(resp).Encode(currentBalance); err != nil {
		logger.Err(err).Msg("failed to encode response")
		server.Error(resp, http.StatusInternalServerError, nil)
		return
	}
}

func (h *balanceHandler) Withdraw(resp http.ResponseWriter, req *http.Request) {
	//goland:noinspection GoUnhandledErrorResult
	defer req.Body.Close()

	ctx, logger := logging.ServiceLogger(req.Context(), balanceHandlerName)
	logger.Info().Msg("handling withdraw")

	var withdraw balance.Withdraw
	if err := json.NewDecoder(req.Body).Decode(&withdraw); err != nil {
		logger.Err(err).Msg("failed to decode request")
		server.Error(resp, http.StatusInternalServerError, nil)
		return
	}

	if err := h.Balance.Withdraw(ctx, AuthorizedUserID(ctx), withdraw); err != nil {
		switch errors.ErrCode(err) {
		case service.ErrOrderWrongOwner:
			server.Error(resp, http.StatusUnprocessableEntity, "invalid order number")
			return
		case service.ErrOrderNotFound:
			server.Error(resp, http.StatusUnprocessableEntity, "invalid order number")
			return
		case service.ErrInsufficientFunds:
			server.Error(resp, http.StatusPaymentRequired, "balance is not enough to withdraw requested sum")
			return
		default:
			server.Error(resp, http.StatusInternalServerError, nil)
		}
	}
}

func (h *balanceHandler) Withdrawals(resp http.ResponseWriter, req *http.Request) {
	//goland:noinspection GoUnhandledErrorResult
	defer req.Body.Close()

	ctx, logger := logging.ServiceLogger(req.Context(), balanceHandlerName)
	logger.Info().Msg("handling withdrawals query")

	withdrawals, err := h.Balance.Withdrawals(ctx, AuthorizedUserID(ctx))
	if err != nil {
		logger.Err(err).Msg("failed to query withdrawals")
		server.Error(resp, http.StatusInternalServerError, nil)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(resp).Encode(withdrawals); err != nil {
		logger.Err(err).Msg("failed to encode response")
		server.Error(resp, http.StatusInternalServerError, nil)
	}
}

func NewBalance(balance service.Balance) http.Handler {
	h := &balanceHandler{balance}

	router := chi.NewRouter()
	router.Get("/", h.Get)
	router.Post("/withdraw", h.Withdraw)
	router.Get("/withdrawals", h.Withdrawals)
	return router
}
