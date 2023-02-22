package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/0x4c6565/p.lee.io/pkg/model"
	"github.com/0x4c6565/p.lee.io/pkg/storage"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
)

type pasteRequest struct {
	Expires int64  `json:"expires"`
	Syntax  string `json:"syntax"`
	Content string `json:"content"`
}

type pasteResponse struct {
	Expires int64  `json:"expires"`
	Syntax  string `json:"syntax"`
	Content string `json:"content"`
	Burnt   bool   `json:"burnt"`
}

type uuidResponse struct {
	ID string `json:"id"`
}

type syntaxResponse struct {
	Label   string   `json:"label"`
	Syntax  string   `json:"syntax"`
	Default bool     `json:"default,omitempty"`
	Aliases []string `json:"aliases,omitempty"`
}

type expiresResponse struct {
	Label   string `json:"label"`
	Default bool   `json:"default,omitempty"`
	Expires int    `json:"expires"`
}

type apiError struct {
	Err  error
	Code int
}

type API struct {
	storage storage.Storage
	config  *Config
	router  *mux.Router
}

func NewAPI(storage storage.Storage, config *Config) *API {
	return &API{
		storage: storage,
		config:  config,
	}
}

func (h *API) Start(ctx context.Context) error {
	log.Info().Msg("Starting API")

	h.router = mux.NewRouter()
	h.router.HandleFunc(`/api/v1/paste/{uuid:\S+}`, h.HandleGetPaste).Methods("GET")
	h.router.HandleFunc("/api/v1/paste", h.HandleCreatePaste).Methods("POST")
	h.router.HandleFunc(`/api/v1/syntax`, h.HandleGetSyntax).Methods("GET")
	h.router.HandleFunc(`/api/v1/expires`, h.HandleGetExpires).Methods("GET")
	// h.router.HandleFunc(`/raw/{uuid:\S+}`, h.HandleGetPasteRaw).Methods("GET")
	h.router.PathPrefix(`/raw/{uuid:[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}}`).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/raw.html")
	})).Methods("GET")
	h.router.PathPrefix(`/{uuid:[0-9a-fA-F]{8}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{4}\-[0-9a-fA-F]{12}}`).Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})).Methods("GET")
	fs := http.FileServer(http.Dir("./static"))
	h.router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fs.ServeHTTP(w, r)
	})).Methods("GET")

	loggedRouter := handlers.LoggingHandler(os.Stdout, h.router)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: loggedRouter,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msgf("HTTP listener failure")
		}
	}()

	log.Info().Msg("API started")
	<-ctx.Done()
	log.Info().Msg("API shutting down..")
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("HTTP listener shutdown failed: %s", err)
	}
	log.Info().Msg("API stopped")

	return nil
}

func (h *API) HandleGetPaste(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	paste, err := h.storage.Get(context.Background(), vars["uuid"])
	if err != nil {
		var notFoundErr *storage.NotFoundError
		if errors.As(err, &notFoundErr) {
			h.handleJSONResponse(w, http.StatusNotFound, "Cannot find paste")
			return
		}

		log.Error().Msgf("Error retrieving paste: %s", err.Error())
		h.handleJSONResponse(w, http.StatusInternalServerError, nil)
		return
	}

	burnt := false
	if paste.Expires == model.PASTE_EXPIRES_BURN {
		err = h.storage.Delete(context.Background(), paste.ID)
		if err != nil {
			log.Error().Msgf("Error burning paste: %s", err.Error())
			h.handleJSONResponse(w, http.StatusInternalServerError, nil)
			return
		}
		burnt = true
	}

	h.handleJSONResponse(w, http.StatusOK, &pasteResponse{
		Expires: paste.Expires,
		Syntax:  paste.Syntax,
		Content: paste.Content,
		Burnt:   burnt,
	})
}

func (h *API) HandleCreatePaste(w http.ResponseWriter, r *http.Request) {
	var req pasteRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)
	if err != nil {
		h.handleJSONResponse(w, http.StatusBadRequest, fmt.Sprintf("Unable to decode JSON payload: %s", err))
		return
	}
	defer r.Body.Close()

	syntaxExists, syntax := h.config.ResolveSyntax(req.Syntax)
	if !syntaxExists {
		h.handleJSONResponse(w, http.StatusBadRequest, "Invalid syntax")
		return
	}

	id, err := h.storage.Add(context.Background(), model.Paste{
		Expires:   req.Expires,
		Timestamp: time.Now().Unix(),
		Syntax:    syntax.Syntax,
		Content:   req.Content,
	})
	if err != nil {
		log.Error().Msgf("Error storing paste: %s", err.Error())
		h.handleJSONResponse(w, http.StatusInternalServerError, nil)
		return
	}

	h.handleJSONResponse(w, http.StatusOK, &uuidResponse{
		ID: id,
	})
}

func (h *API) HandleGetSyntax(w http.ResponseWriter, r *http.Request) {
	var resp []syntaxResponse
	for _, syntax := range h.config.Syntax {
		resp = append(resp, syntaxResponse{
			Label:   syntax.Label,
			Syntax:  syntax.Syntax,
			Default: syntax.Label == h.config.SyntaxDefault,
			Aliases: syntax.Aliases,
		})
	}

	h.handleJSONResponse(w, http.StatusOK, resp)
}

func (h *API) HandleGetExpires(w http.ResponseWriter, r *http.Request) {
	var resp []expiresResponse
	resp = append(resp, expiresResponse{
		Label:   "Never",
		Expires: model.PASTE_EXPIRES_NEVER,
	})
	resp = append(resp, expiresResponse{
		Label:   "Burn after reading",
		Expires: model.PASTE_EXPIRES_BURN,
	})
	for _, expires := range h.config.Expires {
		resp = append(resp, expiresResponse{
			Label:   expires.Label,
			Default: expires.Label == h.config.ExpiresDefault,
			Expires: expires.Expires,
		})
	}

	h.handleJSONResponse(w, http.StatusOK, resp)
}

func (h *API) handleJSONResponse(w http.ResponseWriter, statusCode int, content interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(content)
}
