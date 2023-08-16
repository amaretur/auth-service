package handler

import (
	"github.com/gorilla/mux"

	"github.com/amaretur/auth-service/internal/transport/http/middleware"
)

type handler interface {
	Init(router *mux.Router)
}

type Handler struct {
	router *mux.Router
}

func NewHandler(pathPrefix string) *Handler {

	r := mux.NewRouter().StrictSlash(true).PathPrefix(pathPrefix).Subrouter()

	r.Use(middleware.ApplicationJson)
	r.Use(middleware.ReqId)

	return &Handler{
		router : r,
	}
}

func (h *Handler) Router() *mux.Router {
	return h.router
}

func (h *Handler) Register(sh handler, prefix string) { 
	sh.Init(h.router.PathPrefix(prefix).Subrouter())
}
