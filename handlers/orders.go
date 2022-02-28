package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/zhupanovdm/gophermart/model/order"
	"github.com/zhupanovdm/gophermart/pkg/errors"
	"github.com/zhupanovdm/gophermart/pkg/logging"
	"github.com/zhupanovdm/gophermart/pkg/server"
	"github.com/zhupanovdm/gophermart/pkg/validation"
	"github.com/zhupanovdm/gophermart/service"
)

const ordersHandlerName = "Orders Handler"

type ordersHandler struct {
	service.Orders
}

func (h *ordersHandler) Register(resp http.ResponseWriter, req *http.Request) {
	//goland:noinspection GoUnhandledErrorResult
	defer req.Body.Close()

	ctx, logger := logging.ServiceLogger(req.Context(), ordersHandlerName)
	logger.Info().Msg("handle order registration request")

	data, err := io.ReadAll(req.Body)
	if err != nil {
		logger.Err(err).Msg("can't read request")
		server.Error(resp, http.StatusInternalServerError, nil)
		return
	}

	number := order.Number(data)
	if err := number.Validate(validation.OnlyDigits, validation.Luhn); err != nil {
		logger.Err(err).Msg("invalid order number")
		server.Error(resp, http.StatusUnprocessableEntity, err)
		return
	}

	if err = h.Orders.Register(ctx, number, AuthorizedUserId(ctx)); err != nil {
		switch errors.ErrCode(err) {
		case service.ErrOrderAlreadyRegistered:
			logger.Warn().Msg("order has already registered")
		case service.ErrOrderNumberCollision:
			logger.Err(err).Msg("invalid order number")
			server.Error(resp, http.StatusConflict, err)
		default:
			logger.Err(err).Msg("order registration failed")
			server.Error(resp, http.StatusInternalServerError, nil)
		}
		return
	}

	resp.WriteHeader(http.StatusAccepted)
}

func (h *ordersHandler) GetAll(resp http.ResponseWriter, req *http.Request) {
	//goland:noinspection GoUnhandledErrorResult
	defer req.Body.Close()

	ctx, logger := logging.ServiceLogger(req.Context(), ordersHandlerName)
	logger.Info().Msg("handle client orders query")

	orders, err := h.Orders.GetAll(ctx, AuthorizedUserId(ctx))
	if err != nil {
		logger.Err(err).Msg("orders query failed")
		server.Error(resp, http.StatusInternalServerError, nil)
		return
	}

	if len(orders) == 0 {
		resp.WriteHeader(http.StatusNoContent)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(resp).Encode(orders); err != nil {
		logger.Err(err).Msg("failed to encode response")
		server.Error(resp, http.StatusInternalServerError, nil)
	}
}

func NewOrders(orders service.Orders) http.Handler {
	h := &ordersHandler{orders}

	router := chi.NewRouter()
	router.Get("/", h.GetAll)
	router.Post("/", h.Register)
	return router
}
