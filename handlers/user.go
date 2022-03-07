package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func NewUserHandler(auth http.Handler, orders http.Handler, balance http.Handler) http.Handler {
	router := chi.NewRouter()
	router.Mount("/", auth)
	router.Mount("/balance", balance)
	router.Mount("/orders", orders)
	return router
}
