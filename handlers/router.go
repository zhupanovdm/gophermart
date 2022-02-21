package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func User(auth *Auth) http.Handler {
	router := chi.NewRouter()
	router.Post("/register", auth.Register)
	router.Post("/login", auth.Login)
	return router
}
