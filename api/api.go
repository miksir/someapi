package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"net/http"
	"someAPI/user"
	"strings"
)

type App struct {
	reg    Registry
	logger zerolog.Logger
}

type Registry interface {
	GetUser(ctx context.Context, email string) (user.User, error)
	CreateUser(ctx context.Context, user user.User) error
}

func (a *App) getUser(w http.ResponseWriter, r *http.Request) {
	logger := a.logger.With().Str("request", "getUser").Logger()
	ctx := r.Context()
	path := r.URL.Path
	pathElements := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathElements) < 2 {
		logger.Error().Str("path", r.URL.Path).Msg("malformed URI")
		http.Error(w, "malformed URI", http.StatusBadRequest)
		return
	}

	userFound, err := a.reg.GetUser(ctx, pathElements[1])
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			logger.Warn().Str("path", r.URL.Path).
				Str("user", pathElements[1]).Err(err).Msg("user not found")
			http.Error(w, err.Error(), http.StatusNotFound)
		} else {
			logger.Error().Str("path", r.URL.Path).
				Str("user", pathElements[1]).Err(err).Msg("error requesting user")
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(userFound); err != nil {
		logger.Error().Str("path", r.URL.Path).Str("user", pathElements[1]).
			Interface("user_object", userFound).Err(err).Msg("error to encode to json")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (a *App) createUser(w http.ResponseWriter, r *http.Request) {
	logger := a.logger.With().Str("request", "createUser").Logger()
	var err error
	var userCreate user.User
	err = json.NewDecoder(r.Body).Decode(&userCreate)
	if err != nil {
		logger.Error().Str("path", r.URL.Path).Err(err).Msg("error to decode from json")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = userCreate.Validate()
	if err != nil {
		logger.Error().Str("path", r.URL.Path).Interface("user", userCreate).Msg(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	err = a.reg.CreateUser(ctx, userCreate)
	if err != nil {
		if errors.Is(err, user.ErrUserEmailAlreadyExists) {
			logger.Error().Str("path", r.URL.Path).Err(err).Interface("user", userCreate).Msg("user email already exists")
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		if errors.Is(err, user.ErrUserUUIDAlreadyExists) {
			logger.Error().Str("path", r.URL.Path).Err(err).Interface("user", userCreate).Msg("user UUID already exists")
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		logger.Error().Str("path", r.URL.Path).Err(err).Msg("create user error")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func CreateAPI(logger zerolog.Logger, registry Registry) *App {
	a := &App{reg: registry, logger: logger}
	r := mux.NewRouter()
	r.HandleFunc("/user", a.getUser).Methods("GET")
	r.HandleFunc("/user", a.createUser).Methods("POST")
	http.Handle("/", r)
	return a
}

func (a *App) Run(addr string, handler http.Handler) error {
	return http.ListenAndServe(addr, handler)
}