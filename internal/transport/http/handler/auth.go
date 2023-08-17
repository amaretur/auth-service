package handler

import (
	"time"
	"context"
	"net/http"
	"encoding/json"

	"github.com/gorilla/mux"

	"github.com/amaretur/auth-service/internal/dto"
	"github.com/amaretur/auth-service/internal/validator"

	"github.com/amaretur/auth-service/pkg/log"
)

type Usecase interface {
	SignIn(ctx context.Context, uuid string) (*dto.Tokens, error)
	Refresh(ctx context.Context, tokens *dto.Tokens) (*dto.Tokens, error)
}

type Auth struct {
	usecase	Usecase
	logger	log.Logger
}

func NewAuth(usecase Usecase, logger log.Logger) *Auth {
	return &Auth{
		usecase: usecase,
		logger: logger,
	}
}

func (a *Auth) Init(router *mux.Router) {

	router.HandleFunc("/sign-in", a.Auth).Methods("POST")
	router.HandleFunc("/refresh", a.Refresh).Methods("POST")
}

func (a *Auth) Auth(w http.ResponseWriter, r *http.Request) {

	uuid := r.URL.Query().Get("uuid")

	if err := validator.ValidateUuid(uuid); err != nil {
		Error(w, http.StatusBadRequest, "invalid uuid")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5 * time.Second)
	defer cancel()

	tokens, err := a.usecase.SignIn(ctx, uuid)
	if err != nil {

		code, msg := errToHttpResp(err, defErrHttpMapper)

		logger(r, a.logger, map[string]any{"code": code, "body": msg}).
			Warn(err)

		Error(w, code, msg)
		return
	}

	Response(w, tokens)
}

func (a *Auth) Refresh(w http.ResponseWriter, r *http.Request) {

	var data dto.Tokens

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		Error(w, http.StatusBadRequest, "invalid json structure")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5 * time.Second)
	defer cancel()

	tokens, err := a.usecase.Refresh(ctx, &data)
	if err != nil {

		code, msg := errToHttpResp(err, defErrHttpMapper)

		logger(r, a.logger, map[string]any{"code": code, "body": msg}).
			Warn(err)

		Error(w, code, msg)
		return
	}

	Response(w, tokens)
}
